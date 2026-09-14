package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// typecheckSubErrorRe 匹配 typecheck 消息中的子错误，格式：./path:line:col: message
var typecheckSubErrorRe = regexp.MustCompile(`\./([^:]+):(\d+):(\d+):\s*([^./][^\n]*)`)

// expandTypecheckIssue 将 golangci-lint typecheck 的捆绑消息拆分为多条独立 issue
func expandTypecheckIssue(check *scanModel.IssueInputCheck) []*scanModel.IssueInputCheck {
	if check.Tool != "golangci-lint" || check.RuleKey != "typecheck" {
		return []*scanModel.IssueInputCheck{check}
	}
	matches := typecheckSubErrorRe.FindAllStringSubmatch(check.Message, -1)
	if len(matches) <= 1 {
		return []*scanModel.IssueInputCheck{check}
	}
	result := make([]*scanModel.IssueInputCheck, 0, len(matches))
	for _, m := range matches {
		line, _ := strconv.ParseInt(m[2], 10, 64)
		col, _ := strconv.ParseInt(m[3], 10, 64)
		msg := m[4]
		item := *check
		item.Message = msg
		item.Location.Path = m[1]
		item.Location.Range.StartLine = line
		item.Location.Range.EndLine = line
		item.Location.Range.StartByte = col
		result = append(result, &item)
	}
	return result
}

// AnalyzeTaskUploadReport 分析上传的扫描报告
func (s *ScanTaskService) AnalyzeTaskUploadReport(ctx context.Context, repoID uint, taskID uint, method string, content []byte) (*scanModel.ScanIssueReport, error) {

	var err error
	task := &scanModel.ScanTaskInfo{}
	ppm := &repoModel.PullRequest{}

	if method == "ai" {
		ppm, err = s.pullRequestStore.GetByID(ctx, repoID, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pull request: %w", err)
		}
	} else {
		task, err = s.taskStore.GetTaskInfo(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info: %w", err)
		}
	}

	report := &scanModel.ScanIssueReport{}

	scanMethod := constant.ScanMethod(method)
	var inputIssues []*scanModel.ScanIssue
	switch scanMethod {
	case constant.ScanMethodCheck:
		inputIssues, err = s.analyzeCheckIssues(ctx, content, task, report)
	case constant.ScanMethodSmell:
		inputIssues, err = s.analyzeSmellIssues(ctx, content, task, report)
	case constant.ScanMethodAI:
		inputIssues, err = s.analyzeAIIssues(ctx, content, repoID, int64(ppm.ID), report)
	default:
		return nil, fmt.Errorf("unsupported scan method: %s", method)
	}
	if err != nil {
		return nil, err
	}

	if scanMethod == constant.ScanMethodAI {
		// AI 类型的 issues 直接批量插入
		added, skipped, err := s.saveAIIssues(ctx, inputIssues, ppm)
		if err != nil {
			return nil, err
		}
		// Total 为本次解析出的条数，Add/Skip 为实际入库与跳过的条数
		report.Total = len(inputIssues)
		report.Add = added
		report.Skip = skipped
		return report, nil
	}

	compData, err := s.compareIssues(ctx, task, scanMethod, inputIssues, report)
	if err != nil {
		return nil, err
	}

	// 事务内：保存 issues、绑定关系、分析质量门禁
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if e := s.saveIssues(ctx, compData, task); e != nil {
			return e
		}
		if e := s.bindIssuesByTaskID(ctx, task, compData.All); e != nil {
			return e
		}
		rules, e := s.issueStore.ListByTaskForResult(ctx, taskID, task.PlanID)
		if e != nil {
			return e
		}
		var triggerID int64
		if task.Trigger != nil {
			triggerID = task.Trigger.ID
		}
		result, e := s.analyzeTaskResult(ctx, uint(task.PlanID), uint(triggerID), rules)
		if e != nil {
			return e
		}
		if e = s.taskStore.UpdateResult(ctx, taskID, result); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

// analyzeCheckIssues 解析 check 类型的 issues
func (s *ScanTaskService) analyzeCheckIssues(
	ctx context.Context, content []byte, task *scanModel.ScanTaskInfo, report *scanModel.ScanIssueReport,
) ([]*scanModel.ScanIssue, error) {
	var checks []*scanModel.IssueInputCheck
	if err := json.Unmarshal(content, &checks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal check issues: %w", err)
	}

	report.Total = len(checks)

	ruleKeyMap, err := s.buildRuleKeyMapByTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// 展开 golangci-lint typecheck 捆绑消息
	expanded := make([]*scanModel.IssueInputCheck, 0, len(checks))
	for _, check := range checks {
		expanded = append(expanded, expandTypecheckIssue(check)...)
	}

	var issues []*scanModel.ScanIssue
	for _, check := range expanded {
		inIssue := &scanModel.ScanIssue{
			Title:      check.Message,
			Path:       check.Location.Path,
			StartLine:  check.Location.Range.StartLine,
			ScanMethod: constant.ScanMethodCheck,
		}
		ruleID, ok := ruleKeyMap.GetByTool(check.Tool, check.RuleKey)
		if !ok {
			report.MissingRule++
			continue
		}
		inIssue.RuleID = ruleID
		inIssue.Key = scanModel.GenerateIssueKey(task.RepoID, task.ExecutionRef, check.UniqueKey())
		if err := inIssue.SetPayload(check.IssuePayloadCheck); err != nil {
			return nil, err
		}
		issues = append(issues, inIssue)
	}
	return issues, nil
}

// analyzeSmellIssues 解析 smell 类型的 issues
func (s *ScanTaskService) analyzeSmellIssues(
	ctx context.Context, content []byte, task *scanModel.ScanTaskInfo, report *scanModel.ScanIssueReport,
) ([]*scanModel.ScanIssue, error) {
	var smells []*scanModel.IssueInputSmell
	if err := json.Unmarshal(content, &smells); err != nil {
		return nil, fmt.Errorf("failed to unmarshal smell issues: %w", err)
	}

	report.Total = len(smells)

	ruleKeyMap, err := s.buildRuleKeyMapByTask(ctx, task)
	if err != nil {
		return nil, err
	}

	var issues []*scanModel.ScanIssue
	for _, smell := range smells {
		inIssue := &scanModel.ScanIssue{
			Title:      smell.Message,
			Path:       smell.Location.Path,
			StartLine:  smell.Location.Range.StartLine,
			ScanMethod: constant.ScanMethodSmell,
		}
		ruleID, ok := ruleKeyMap.GetByTool(smell.Tool, smell.RuleKey)
		if !ok {
			report.MissingRule++
			continue
		}
		inIssue.RuleID = ruleID
		inIssue.Key = scanModel.GenerateIssueKey(task.RepoID, task.ExecutionRef, smell.UniqueKey())
		if err := inIssue.SetPayload(smell.IssuePayloadSmell); err != nil {
			return nil, err
		}
		issues = append(issues, inIssue)
	}
	return issues, nil
}

// analyzeAIIssues 解析 ai 类型的 issues
func (s *ScanTaskService) analyzeAIIssues(
	ctx context.Context, content []byte, repoID uint, ppmID int64, report *scanModel.ScanIssueReport,
) ([]*scanModel.ScanIssue, error) {
	var aiReport scanModel.AIReportResponse
	if err := json.Unmarshal(content, &aiReport); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI report: %w", err)
	}

	// ocr 上报的 status 有 complete/partial/skipped/failed 等取值，其中 partial 表示部分
	// 文件评审失败但仍有可用意见（ocr 退出码为 0）。这里只以意见数量为准，避免上游新增
	// 状态值时整份报告被静默丢弃。
	if len(aiReport.Comments) == 0 {
		logger.GetLogger(ctx).Infof("[ai-review] no comments in report, skip: status=%s", aiReport.Status)
		return nil, nil
	}

	report.Total = len(aiReport.Comments)

	var issues []*scanModel.ScanIssue
	for _, comment := range aiReport.Comments {
		comment.PpmID = ppmID
		title := comment.Content
		if runes := []rune(title); len(runes) > 30 {
			title = string(runes[:30])
		}
		inIssue := &scanModel.ScanIssue{
			Title:      title,
			Content:    comment.Content,
			Path:       comment.Path,
			StartLine:  comment.StartLine,
			EndLine:    comment.EndLine,
			ScanMethod: constant.ScanMethodAI,
			OldCode:    comment.ExistingCode,
			NewCode:    comment.SuggestionCode,
			Category:   constant.AiScanIssueCategory(comment.Category),
			Severity:   constant.AiScanIssueSeverity(comment.Severity),
		}
		inIssue.Key = scanModel.GenerateIssueKey(int64(repoID), "", comment.UniqueKey())
		if err := inIssue.SetPayload(scanModel.IssuePayloadAI{
			SuggestionCode: comment.SuggestionCode,
			ExistingCode:   comment.ExistingCode,
		}); err != nil {
			return nil, err
		}
		issues = append(issues, inIssue)
	}
	return issues, nil
}

// buildRuleKeyMapByTask 根据任务构建规则 key 映射
func (s *ScanTaskService) buildRuleKeyMapByTask(ctx context.Context, task *scanModel.ScanTaskInfo) (scanModel.RulesMap, error) {
	var solutionIDs []int64

	// 如果有 trigger，获取 trigger 关联的 solutionIDs
	if task.Trigger != nil && task.Trigger.ID != 0 {
		trigger, err := s.planStore.GetByID(ctx, uint(task.PlanID))
		if err != nil {
			return nil, fmt.Errorf("failed to find plan: %w", err)
		}
		_ = trigger
		// 使用 plan 的 solutionIDs
	}

	// 获取 plan 关联的 solutionIDs
	if len(solutionIDs) == 0 {
		ids, err := s.planStore.GetSolutionIDList(ctx, uint(task.PlanID))
		if err != nil {
			return nil, fmt.Errorf("failed to get solution ids: %w", err)
		}
		solutionIDs = ids
	}
	if len(solutionIDs) == 0 {
		return nil, fmt.Errorf("no solutions found for plan")
	}

	// 遍历 solution → ruleset → rule，构建映射
	m := make(scanModel.RulesMap)
	for _, slnID := range solutionIDs {
		solution, err := s.solutionStore.GetByID(ctx, uint(slnID))
		if err != nil || solution == nil || solution.Status != constant.ScanRuleStatusEnabled {
			continue
		}
		rulesets, err := s.solutionStore.ListRulesets(ctx, uint(slnID))
		if err != nil {
			return nil, fmt.Errorf("failed to list rulesets: %w", err)
		}
		for _, rs := range rulesets {
			if rs.Status != constant.ScanRuleStatusEnabled {
				continue
			}
			ruleIDs, err := s.ruleSetStore.GetRuleIDList(ctx, uint(rs.ID))
			if err != nil {
				return nil, fmt.Errorf("failed to get rule ids: %w", err)
			}
			if len(ruleIDs) == 0 {
				continue
			}
			rules, err := s.ruleStore.ListByIDs(ctx, ruleIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to list rules: %w", err)
			}
			for _, rule := range rules {
				if rule.Status != constant.ScanRuleStatusEnabled {
					continue
				}
				indexKey := fmt.Sprintf("%s:%s", rule.Lang.LowerString(), rule.RuleKey)
				m[indexKey] = rule.ID
			}
		}
	}
	return m, nil
}

// compareIssues 将本次上传的 issues 与数据库中已有的 issues 做 diff
func (s *ScanTaskService) compareIssues(
	ctx context.Context, task *scanModel.ScanTaskInfo, method constant.ScanMethod,
	inputs []*scanModel.ScanIssue, report *scanModel.ScanIssueReport,
) (*scanModel.ScanIssueComparedData, error) {
	existIssues, err := s.issueStore.ListKeysByTask(ctx, task.RepoID, task.ExecutionRef, uint(task.ID), method)
	if err != nil {
		return nil, err
	}

	var toUpdateIDs []int64
	var toAddItems []*scanModel.ScanIssue
	var allIssues []*scanModel.ScanIssue
	var expiredIgnoreIDs []int64

	listBMap := make(map[string]*scanModel.ScanIssue)
	for _, item := range inputs {
		allIssues = append(allIssues, item)
		if _, exist := listBMap[item.Key]; exist {
			report.Conflict++
			continue
		}
		listBMap[item.Key] = item
	}

	listAMap := make(map[string]bool)
	for _, item := range existIssues {
		listAMap[item.Key] = true
	}

	now := time.Now().Unix()
	for _, item := range existIssues {
		if _, exist := listBMap[item.Key]; exist {
			// ignore 状态：永久忽略（-1）直接跳过；未过期跳过；过期则重置
			if item.Status == constant.ScanIssueStatusIgnore {
				if item.IgnoredDate != nil && (now <= *item.IgnoredDate || *item.IgnoredDate == -1) {
					report.Skip++
					continue
				}
				expiredIgnoreIDs = append(expiredIgnoreIDs, item.ID)
			}
			if item.UpdatedByTaskID >= task.ID {
				report.Skip++
				continue
			}
			toUpdateIDs = append(toUpdateIDs, item.ID)
		}
	}

	for key, item := range listBMap {
		if _, exists := listAMap[key]; !exists {
			toAddItems = append(toAddItems, item)
		}
	}

	if len(expiredIgnoreIDs) > 0 {
		if err := s.issueStore.ResetExpiredIgnore(ctx, expiredIgnoreIDs, req.GetOperator(ctx)); err != nil {
			return nil, err
		}
	}

	report.Update = len(toUpdateIDs)
	report.Add = len(toAddItems)

	return &scanModel.ScanIssueComparedData{
		Added:   toAddItems,
		Updated: toUpdateIDs,
		All:     allIssues,
	}, nil
}

// saveIssues 保存 issue 变更（增/改/删）
func (s *ScanTaskService) saveIssues(ctx context.Context, comp *scanModel.ScanIssueComparedData, task *scanModel.ScanTaskInfo) error {
	if len(comp.Updated) > 0 {
		pages := paginateIndex(len(comp.Updated), 1000)
		for _, p := range pages {
			if err := s.issueStore.UpdateByTaskID(ctx, uint(task.ID), comp.Updated[p[0]:p[1]], req.GetOperator(ctx)); err != nil {
				return err
			}
		}
	}

	if len(comp.Added) > 0 {
		emptyIgnoredDate := int64(0)
		storeIssues := make([]*scanStore.ScanIssue, 0, len(comp.Added))
		for _, item := range comp.Added {
			item.RepoID = task.RepoID
			item.RepoBranch = task.ExecutionRef
			item.CreatedByTaskID = task.ID
			item.UpdatedByTaskID = task.ID
			item.Status = constant.ScanIssueStatusWait
			storeIssues = append(storeIssues, &scanStore.ScanIssue{
				Key:             item.Key,
				RuleID:          item.RuleID,
				Title:           item.Title,
				Path:            item.Path,
				StartLine:       item.StartLine,
				RepoID:          item.RepoID,
				RepoBranch:      item.RepoBranch,
				CreatedByTaskID: item.CreatedByTaskID,
				UpdatedByTaskID: item.UpdatedByTaskID,
				Status:          item.Status,
				ScanMethod:      item.ScanMethod,
				Resolution:      item.Resolution,
				Category:        item.Category.String(),
				Severity:        item.Severity.String(),
				Payload:         item.Payload,
				CreatedBy:       task.CreatedBy,
				IgnoredDate:     &emptyIgnoredDate,
			})
		}
		pages := paginateIndex(len(storeIssues), 1000)
		for _, p := range pages {
			if err := s.issueStore.BatchCreate(ctx, storeIssues[p[0]:p[1]]); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveAIIssues 直接批量插入 AI 类型的 issues，返回实际插入与跳过的条数（跳过的 key 已存在）
func (s *ScanTaskService) saveAIIssues(ctx context.Context, issues []*scanModel.ScanIssue, ppm *repoModel.PullRequest) (int, int, error) {
	if len(issues) == 0 {
		return 0, 0, nil
	}

	rawIssues := make([]*scanStore.ScanIssue, 0, len(issues))
	emptyResolution := constant.ScanIssueResolutionNull
	emptyIgnoredDate := int64(0)
	for _, item := range issues {
		item.PpmID = int64(ppm.ID)
		item.RepoID = int64(ppm.RepoID)
		item.RepoBranch = ppm.SourceBranch
		item.Status = constant.ScanIssueStatusWait
		item.ScanMethod = constant.ScanMethodAI
		rawIssues = append(rawIssues, &scanStore.ScanIssue{
			Key:         item.Key,
			Title:       item.Title,
			Content:     item.Content,
			Path:        item.Path,
			StartLine:   item.StartLine,
			EndLine:     item.EndLine,
			OldCode:     item.OldCode,
			NewCode:     item.NewCode,
			PpmID:       item.PpmID,
			RepoID:      item.RepoID,
			Status:      item.Status,
			ScanMethod:  item.ScanMethod,
			Resolution:  &emptyResolution,
			IgnoredDate: &emptyIgnoredDate,
			Payload:     item.Payload,
			CreatedBy:   ppm.CreatedBy,
			Category:    string(item.Category),
			Severity:    string(item.Severity),
		})
	}

	// 批内去重：同一份报告里可能存在 path/startLine/endLine 相同的 comment，key 会重复，
	// 不去重会导致整批插入触发唯一键冲突
	keys := make([]string, 0, len(rawIssues))
	seen := make(map[string]struct{}, len(rawIssues))
	deduped := make([]*scanStore.ScanIssue, 0, len(rawIssues))
	skipped := 0
	for _, item := range rawIssues {
		if _, ok := seen[item.Key]; ok {
			skipped++
			continue
		}
		seen[item.Key] = struct{}{}
		keys = append(keys, item.Key)
		deduped = append(deduped, item)
	}

	// 预校验：已存在的 issue 不重复插入
	existingKeys, err := s.issueStore.ListExistingKeys(ctx, keys)
	if err != nil {
		return 0, 0, err
	}

	storeIssues := deduped
	if len(existingKeys) > 0 {
		existingSet := make(map[string]struct{}, len(existingKeys))
		for _, k := range existingKeys {
			existingSet[k] = struct{}{}
		}
		filterIssues := deduped[:0]
		for _, item := range deduped {
			if _, ok := existingSet[item.Key]; ok {
				skipped++
				continue
			}
			filterIssues = append(filterIssues, item)
		}
		storeIssues = filterIssues
	}
	if len(storeIssues) == 0 {
		return 0, skipped, nil
	}

	added := 0
	pages := paginateIndex(len(storeIssues), 1000)
	for _, p := range pages {
		page := storeIssues[p[0]:p[1]]
		if err := s.issueStore.BatchCreate(ctx, page); err == nil {
			added += len(page)
			continue
		} else if !errors.Is(err, errors.ErrDuplicate) {
			return 0, 0, err
		}

		// 并发插入仍可能冲突：降级为逐条插入，只忽略重复的 issue，整批不再丢弃
		for _, item := range page {
			err := s.issueStore.Create(ctx, item)
			if err == nil {
				added++
				continue
			}
			if errors.Is(err, errors.ErrDuplicate) {
				skipped++
				continue
			}
			return 0, 0, err
		}
	}
	return added, skipped, nil
}

// bindIssuesByTaskID 绑定 issues 到 task，并更新 task 的 issue 数量
func (s *ScanTaskService) bindIssuesByTaskID(ctx context.Context, task *scanModel.ScanTaskInfo, issues []*scanModel.ScanIssue) error {
	issueKeys := make([]string, 0, len(issues))
	seen := make(map[string]struct{})
	for _, issue := range issues {
		if _, ok := seen[issue.Key]; !ok {
			seen[issue.Key] = struct{}{}
			issueKeys = append(issueKeys, issue.Key)
		}
	}
	sort.Strings(issueKeys)

	if len(issueKeys) > 0 {
		pages := paginateIndex(len(issueKeys), 1000)
		for _, p := range pages {
			if err := s.issueStore.CreateIssueBindByTaskID(ctx, uint(task.ID), issueKeys[p[0]:p[1]]); err != nil {
				return err
			}
		}
	}
	if err := s.taskStore.UpdateIssueNumber(ctx, uint(task.ID), int64(len(issueKeys))); err != nil {
		return err
	}
	return nil
}

// analyzeTaskResult 根据质量门禁条件判定任务结果
func (s *ScanTaskService) analyzeTaskResult(
	ctx context.Context, planID, triggerID uint, rules []*scanModel.Rule,
) (constant.ScanTaskResult, error) {
	conditions, err := s.conditionStore.ListByPlanID(ctx, planID)
	if err != nil {
		return "", err
	}

	if len(conditions) == 0 {
		return constant.ScanTaskResultPass, nil
	}

	counter := make([]int64, len(conditions))
	for _, rule := range rules {
		for idx, cond := range conditions {
			if cond.Type != "" && rule.Type != cond.Type {
				continue
			}
			if cond.Priority != "" && rule.Priority != cond.Priority {
				continue
			}
			counter[idx]++
		}
	}

	totalNum := int64(len(rules))
	result := constant.ScanTaskResultPass
	for idx, cond := range conditions {
		if counter[idx] == 0 {
			continue
		}
		var val int64
		switch cond.Unit {
		case constant.ScanRuleUnitPercent:
			if totalNum > 0 {
				val = (counter[idx] * 100) / totalNum
			}
		case constant.ScanRuleUnitCount:
			val = counter[idx]
		default:
			val = counter[idx]
		}
		if val > cond.Threshold {
			result = constant.ScanTaskResultNoPass
			break
		}
	}
	return result, nil
}

// paginateIndex 将长度为 length 的切片按 pageSize 分页，返回每页的 [start, end) 索引
func paginateIndex(length, pageSize int) [][2]int {
	var pages [][2]int
	for i := 0; i < length; i += pageSize {
		end := i + pageSize
		if end > length {
			end = length
		}
		pages = append(pages, [2]int{i, end})
	}
	return pages
}

// GetIssueInfo 获取单个扫描问题详情
func (s *ScanTaskService) GetIssueInfo(ctx context.Context, issueID uint) (*scanModel.ScanIssueInfo, error) {
	return s.issueStore.Get(ctx, issueID)
}

// ListIssueInfos 查询扫描问题列表
func (s *ScanTaskService) ListIssueInfos(ctx context.Context, filter scanModel.ScanIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error) {
	return s.issueStore.List(ctx, filter)
}

// ListIssueInfosByIDs 根据 ID 列表查询扫描问题
func (s *ScanTaskService) ListIssueInfosByIDs(ctx context.Context, ids []int64) ([]*scanModel.ScanIssueInfo, error) {
	return s.issueStore.ListByIDs(ctx, ids)
}

// UpdateIssuesStatus 批量更新扫描问题状态
func (s *ScanTaskService) UpdateIssuesStatus(ctx context.Context, ids []int64, status constant.ScanIssueStatus) error {
	return s.issueStore.UpdateStatus(ctx, ids, status, req.GetOperator(ctx))
}

// SetIssuesStatus 批量更新扫描问题状态和/或解决方案（事务）
func (s *ScanTaskService) SetIssuesStatus(ctx context.Context, in scanModel.ScanIssueUpdateStatusReq) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if in.Status != nil {
			if err := s.issueStore.UpdateStatus(ctx, in.IDs, *in.Status, req.GetOperator(ctx)); err != nil {
				return err
			}
			if err := s.issueStore.UpdateTimestamps(ctx, in, req.GetOperator(ctx)); err != nil {
				return err
			}
		}
		if in.Resolution != nil {
			var resolution *constant.ScanIssueResolution
			if *in.Resolution != constant.ScanIssueResolutionNull {
				resolution = in.Resolution
			}
			if err := s.issueStore.UpdateResolution(ctx, in.IDs, resolution, req.GetOperator(ctx)); err != nil {
				return err
			}
		}
		return nil
	})
}
