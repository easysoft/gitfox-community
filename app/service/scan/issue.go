package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/constant"
	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/database"
	gitAPI "github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// typecheckSubErrorRe 匹配 typecheck 消息中的一条子错误（<文件>:<行>:<列>: <错误>）。
// 由调用方逐行匹配：golangci-lint 对 root 包的文件输出 `./main.go:1:2:`、对子目录包
// 输出 `pkg/a.go:1:2:`（无 ./ 前缀），两种形态都要认；逐行匹配也避免单条大正则把
// 下一行的文件路径吞进 message。
var typecheckSubErrorRe = regexp.MustCompile(`^([^:\s]+):(\d+):(\d+):\s*(.*)$`)

// expandTypecheckIssue 将 golangci-lint typecheck 的捆绑消息拆分为多条独立 issue。
//
// typecheck 把同一个包内的多个编译错误塞进一条 message（首行 `# 包名`，其后每行
// `<文件>:<行>:<列>: <错误>`），location 却只锚在包内首个文件的第 1 行。不拆开的话
// 整包错误只算一条 issue 且位置错误——按行号做增量过滤时（codescan 插件按 git diff
// 新增行过滤）还会因为锚点行不在改动范围内而整条丢失。
//
// 子错误沿用原 issue 的 tool/ruleKey/等级/类别，位置改写为自身的路径/行/列；
// snippet、snippetWithContext 与 location.commit 属于锚点位置，一并清空避免张冠李戴。
// codescan 插件上报前已展开的报告不会再命中 typecheckSubErrorRe，不会被二次拆分。
func expandTypecheckIssue(check *scanModel.IssueInputCheck) []*scanModel.IssueInputCheck {
	if check.Tool != "golangci-lint" || check.RuleKey != "typecheck" {
		return []*scanModel.IssueInputCheck{check}
	}

	type subError struct {
		path    string
		line    int64
		col     int64
		message string
	}
	var subs []subError
	for _, line := range strings.Split(check.Message, "\n") {
		m := typecheckSubErrorRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		lineNo, lineErr := strconv.ParseInt(m[2], 10, 64)
		col, colErr := strconv.ParseInt(m[3], 10, 64)
		if lineErr != nil || colErr != nil || lineNo <= 0 {
			continue
		}
		path := strings.TrimPrefix(m[1], "./")
		// 绝对路径（仓库自带 .golangci.yml 配 path-mode: abs）与库内相对路径口径
		// 不一致：跳过该子错误，整条 issue 保持原样
		if filepath.IsAbs(path) {
			continue
		}
		subs = append(subs, subError{
			path:    path,
			line:    lineNo,
			col:     col,
			message: strings.TrimSpace(m[4]),
		})
	}
	if len(subs) == 0 {
		return []*scanModel.IssueInputCheck{check}
	}

	result := make([]*scanModel.IssueInputCheck, 0, len(subs))
	for _, s := range subs {
		item := *check // 值拷贝，&item 各条独立
		item.Message = s.message
		item.Snippet = ""
		item.SnippetWithContext = ""
		// 重建 location：锚点的 range 与 commit 一律清空，只留子错误自身的位置
		item.Location = scanModel.IssueLocation{Path: s.path}
		item.Location.Range.StartLine = s.line
		item.Location.Range.EndLine = s.line
		item.Location.Range.StartColumn = s.col
		result = append(result, &item)
	}
	return result
}

// AnalyzeTaskUploadReport 分析上传的扫描报告
// taskID：codescan 场景为扫描任务 ID；AI 评审场景语义为 ppmID，可为 0（无 MR，由 executionID 归属执行）
func (s *ScanTaskService) AnalyzeTaskUploadReport(ctx context.Context, repoID, taskID uint, executionID int64, method string, content []byte) (*scanModel.ScanIssueReport, error) {

	var err error
	task := &scanModel.ScanTaskInfo{}
	ppm := &repoModel.PullRequest{}
	var execution *executionModel.Execution

	if method == "ai" {
		if taskID == 0 {
			// ppmID=0 表示无 MR（推送/内嵌流水线触发）：报告归属到 executionID 对应的执行，
			// 与 codescan create-by-runner 共用同一条执行 ID 链路
			if executionID <= 0 {
				return nil, fmt.Errorf("executionID is required when ppmID is 0")
			}
			execution, err = s.executionStore.GetByID(ctx, uint(executionID))
			if err != nil {
				return nil, fmt.Errorf("failed to get execution by ID %d: %w", executionID, err)
			}
		} else {
			ppm, err = s.pullRequestStore.GetByID(ctx, repoID, taskID)
			if err != nil {
				return nil, fmt.Errorf("failed to get pull request: %w", err)
			}
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
		// 唯一键作用域：有 MR 用 ppmID、无 MR 用 executionID。两者来自互相独立的自增表
		// （ops_ppm.id / ops_pipeline_executions.id），同一仓库内 ID 可能相同，故带前缀区分
		// （key = ai/{ppm|exec}/{scopeID}/{path}/{startLine}/{endLine}）
		scope := scanModel.AIScopePPM(int64(ppm.ID))
		if execution != nil {
			scope = scanModel.AIScopeExecution(int64(execution.ID))
		}
		inputIssues, err = s.analyzeAIIssues(ctx, content, repoID, scope, report)
	default:
		return nil, fmt.Errorf("unsupported scan method: %s", method)
	}
	if err != nil {
		return nil, err
	}

	if scanMethod == constant.ScanMethodAI {
		// AI 类型的 issues 直接批量插入
		var saveErr error
		var added, skipped int
		if execution != nil {
			// 无 MR 场景：以 executionID 查出的执行数据填充（createdBy/ref 等）
			added, skipped, saveErr = s.saveAIIssuesForExecution(ctx, inputIssues, execution, int64(repoID))
		} else {
			added, skipped, saveErr = s.saveAIIssues(ctx, inputIssues, ppm)
		}
		if saveErr != nil {
			return nil, saveErr
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

	ruleKeyMap, err := s.buildRuleKeyMapByTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// 展开 golangci-lint typecheck 捆绑消息（一条 message 里含多个子错误）
	expanded := make([]*scanModel.IssueInputCheck, 0, len(checks))
	for _, check := range checks {
		expanded = append(expanded, expandTypecheckIssue(check)...)
	}

	// Total 取展开后的条数：展开出的子错误各自入库，与 MissingRule 等统计口径一致
	report.Total = len(expanded)

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
// scope 为 issue 唯一键的作用域标识，由 scanModel.AIScopePPM（有 MR）/ AIScopeExecution
// （无 MR）构造，两者 ID 空间独立、不可混用
func (s *ScanTaskService) analyzeAIIssues(
	ctx context.Context, content []byte, repoID uint, scope string, report *scanModel.ScanIssueReport,
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
		inIssue.Key = scanModel.GenerateIssueKey(int64(repoID), "", comment.UniqueKey(scope))
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
	for _, item := range issues {
		item.PpmID = int64(ppm.ID)
		item.RepoID = int64(ppm.RepoID)
		// repoBranch 与另外两个写入点（codescan 的 task.ExecutionRef、无 MR 的 execution.Ref）
		// 统一为完整 ref：ppm.SourceBranch 是短分支名，直接落库会让同一列的两种格式并存，
		// 按分支等值过滤（store/scan/issue.go 的 ListKeysByTask、issue_tree.go 的分支过滤）
		// 和按 repoBranch 分组时命中不到
		item.RepoBranch = gitAPI.GetReferenceFromBranchName(ppm.SourceBranch)
	}
	return s.insertAIIssues(ctx, issues, ppm.CreatedBy)
}

// saveAIIssuesForExecution 无 MR（ppmID=0）场景：AI issues 归属到 executionID 对应的执行，
// 归属数据（createdBy/ref/repoID）由 execution 记录填充，与 ppm 关联的字段全部置 0/留空。
// execution.Ref 本身即完整 ref（refs/heads/xxx、refs/pullreq/N/head），与 saveAIIssues 的
// 归一化口径一致，这里不做转换
func (s *ScanTaskService) saveAIIssuesForExecution(ctx context.Context, issues []*scanModel.ScanIssue, execution *executionModel.Execution, repoID int64) (int, int, error) {
	if len(issues) == 0 {
		return 0, 0, nil
	}
	for _, item := range issues {
		item.PpmID = 0
		item.ExecutionID = int64(execution.ID)
		item.RepoID = repoID
		item.RepoBranch = execution.Ref
	}
	return s.insertAIIssues(ctx, issues, execution.CreatedBy)
}

// insertAIIssues 直接批量插入 AI 类型的 issues（按 key 去重）
// 返回实际插入与跳过的条数（跳过的 key 已存在）
func (s *ScanTaskService) insertAIIssues(ctx context.Context, issues []*scanModel.ScanIssue, createdBy string) (int, int, error) {
	if len(issues) == 0 {
		return 0, 0, nil
	}

	rawIssues := make([]*scanStore.ScanIssue, 0, len(issues))
	emptyResolution := constant.ScanIssueResolutionNull
	emptyIgnoredDate := int64(0)
	for _, item := range issues {
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
			ExecutionID: item.ExecutionID,
			RepoID:      item.RepoID,
			RepoBranch:  item.RepoBranch,
			Status:      item.Status,
			ScanMethod:  item.ScanMethod,
			Resolution:  &emptyResolution,
			IgnoredDate: &emptyIgnoredDate,
			Payload:     item.Payload,
			CreatedBy:   createdBy,
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
