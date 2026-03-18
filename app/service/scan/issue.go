package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
)

// AnalyzeTaskUploadReport 分析上传的扫描报告
func (s *ScanTaskService) AnalyzeTaskUploadReport(
	ctx context.Context, taskID uint, method string, content []byte,
) (*scanModel.ScanIssueReport, error) {
	task, err := s.taskStore.GetTaskInfo(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}

	report := &scanModel.ScanIssueReport{}

	scanMethod := constant.ScanMethod(method)
	var inputIssues []*scanModel.ScanIssue
	switch scanMethod {
	case constant.ScanMethodCheck:
		inputIssues, err = s.analyzeCheckIssues(ctx, content, task, report)
	case constant.ScanMethodSmell:
		inputIssues, err = s.analyzeSmellIssues(ctx, content, task, report)
	default:
		return nil, fmt.Errorf("unsupported scan method: %s", method)
	}
	if err != nil {
		return nil, err
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

	var issues []*scanModel.ScanIssue
	for _, check := range checks {
		inIssue := &scanModel.ScanIssue{
			Message:    check.Message,
			Path:       check.Location.Path,
			Line:       check.Location.Range.StartLine,
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
			Message:    smell.Message,
			Path:       smell.Location.Path,
			Line:       smell.Location.Range.StartLine,
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
				m[rule.RuleKey] = rule.ID
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

	var toDeleteIDs []int64
	var toUpdateIDs []int64
	var toAddItems []*scanModel.ScanIssue
	var allIssues []*scanModel.ScanIssue

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

	for _, item := range existIssues {
		if _, exist := listBMap[item.Key]; !exist {
			toDeleteIDs = append(toDeleteIDs, item.ID)
		} else {
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

	report.Resolve = len(toDeleteIDs)
	report.Update = len(toUpdateIDs)
	report.Add = len(toAddItems)

	return &scanModel.ScanIssueComparedData{
		Added:   toAddItems,
		Updated: toUpdateIDs,
		Removed: toDeleteIDs,
		All:     allIssues,
	}, nil
}

// saveIssues 保存 issue 变更（增/改/删）
func (s *ScanTaskService) saveIssues(ctx context.Context, comp *scanModel.ScanIssueComparedData, task *scanModel.ScanTaskInfo) error {
	if len(comp.Updated) > 0 {
		pages := paginateIndex(len(comp.Updated), 1000)
		for _, p := range pages {
			if err := s.issueStore.UpdateByTaskID(ctx, uint(task.ID), comp.Updated[p[0]:p[1]]); err != nil {
				return err
			}
		}
	}

	if len(comp.Removed) > 0 {
		pages := paginateIndex(len(comp.Removed), 1000)
		for _, p := range pages {
			if err := s.issueStore.DisappearedByTask(ctx, uint(task.ID), comp.Removed[p[0]:p[1]]); err != nil {
				return err
			}
		}
	}

	if len(comp.Added) > 0 {
		now := time.Now().UnixMilli()
		for _, item := range comp.Added {
			item.RepoID = task.RepoID
			item.RepoBranch = task.ExecutionRef
			item.Created = now
			item.Updated = now
			item.CreatedByTaskID = task.ID
			item.UpdatedByTaskID = task.ID
			item.Status = constant.ScanIssueStatusWait
			issue := &scanStore.ScanIssue{
				Key:             item.Key,
				RuleID:          item.RuleID,
				Message:         item.Message,
				Path:            item.Path,
				Line:            item.Line,
				RepoID:          item.RepoID,
				RepoBranch:      item.RepoBranch,
				Created:         item.Created,
				Updated:         item.Updated,
				CreatedByTaskID: item.CreatedByTaskID,
				UpdatedByTaskID: item.UpdatedByTaskID,
				Status:          item.Status,
				ScanMethod:      item.ScanMethod,
				Payload:         item.Payload,
			}
			if err := s.issueStore.Create(ctx, issue); err != nil {
				return err
			}
		}
	}
	return nil
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
	conditions, err := s.conditionStore.List(ctx, triggerID)
	if err != nil {
		return "", err
	}

	if len(conditions) == 0 {
		// 如果 trigger 级别没有条件，尝试获取 plan 级别条件
		if triggerID > 0 {
			conditions, err = s.conditionStore.List(ctx, 0)
			if err != nil {
				return "", err
			}
			if len(conditions) == 0 {
				return constant.ScanTaskResultPass, nil
			}
		} else {
			return constant.ScanTaskResultPass, nil
		}
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
