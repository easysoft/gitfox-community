// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	pipelineService "github.com/easysoft/gitfox-community/app/service/pipeline"
	executionStore "github.com/easysoft/gitfox-community/app/store/execution"
	pipelineStore "github.com/easysoft/gitfox-community/app/store/pipeline"
	pullRequestStore "github.com/easysoft/gitfox-community/app/store/pullrequest"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/database/dbtx"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

const (
	_codeScanPipelineName = "_codescan"
	_codeScanLabel        = "codescan"
	// _codeScanCsharpLabel  = "codescan_csharp"
	_codeScanDescription = "inner codescan pipeline"

	JobTypeCodeScan        = "gitfox:codescan"
	JobMaxDurationCodeScan = 2 * time.Hour
)

// ScanTaskInput 创建扫描任务的输入参数
type ScanTaskInput struct {
	Plan    *scanModel.ScanPlan
	Repo    *repoModel.Repo
	Trigger *scanModel.ScanTaskTrigger
	Hook    *triggerer.Hook
}

// JobCodeScanIdentifier 定时扫描任务标识
type JobCodeScanIdentifier struct {
	RepoID    int64 `json:"repo_id"`
	PlanID    int64 `json:"plan_id"`
	TriggerID int64 `json:"trigger_id"`
}

func (j *JobCodeScanIdentifier) String() string {
	return fmt.Sprintf("%s:%d:%d:%d", JobTypeCodeScan, j.RepoID, j.PlanID, j.TriggerID)
}

var _ IScanTask = (*ScanTaskService)(nil)

type ScanTaskService struct {
	config           *config.Config
	taskStore        scanStore.IScanTask
	planStore        scanStore.IScanPlan
	solutionStore    scanStore.ISolution
	issueStore       scanStore.IScanIssue
	conditionStore   scanStore.IScanCondition
	ruleSetStore     scanStore.IRuleSet
	ruleStore        scanStore.IRule
	repoStore        repoStore.Interface
	pipelineStore    pipelineStore.Interface
	pullRequestStore pullRequestStore.Interface
	executionStore   executionStore.Interface
	pipelineSvc      pipelineService.Interface
	triggerSvc       triggerer.Triggerer
	cronScheduler    CronScheduler
	commitLister     CommitLister
	tx               dbtx.TxRunner
}

func NewScanTaskService(
	config *config.Config,
	taskStore scanStore.IScanTask,
	planStore scanStore.IScanPlan,
	solutionStore scanStore.ISolution,
	issueStore scanStore.IScanIssue,
	conditionStore scanStore.IScanCondition,
	ruleSetStore scanStore.IRuleSet,
	ruleStore scanStore.IRule,
	repoStore repoStore.Interface,
	pipelineStore pipelineStore.Interface,
	pullRequestStore pullRequestStore.Interface,
	executionStore executionStore.Interface,
	pipelineService pipelineService.Interface,
	triggerSvc triggerer.Triggerer,
	cronScheduler CronScheduler,
	commitLister CommitLister,
	tx dbtx.TxRunner,
) *ScanTaskService {
	return &ScanTaskService{
		config:           config,
		taskStore:        taskStore,
		planStore:        planStore,
		solutionStore:    solutionStore,
		issueStore:       issueStore,
		conditionStore:   conditionStore,
		ruleSetStore:     ruleSetStore,
		ruleStore:        ruleStore,
		repoStore:        repoStore,
		pipelineStore:    pipelineStore,
		pullRequestStore: pullRequestStore,
		executionStore:   executionStore,
		pipelineSvc:      pipelineService,
		triggerSvc:       triggerSvc,
		cronScheduler:    cronScheduler,
		commitLister:     commitLister,
		tx:               tx,
	}
}

// Create 创建扫描任务
func (s *ScanTaskService) Create(ctx context.Context, in *ScanTaskInput) (*scanModel.ScanTask, error) {
	// 计算序号
	planNumber, err := s.taskStore.NextPlanNumber(ctx, uint64(in.Plan.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get next plan number: %w", err)
	}
	repoNumber, err := s.taskStore.NextRepoNumber(ctx, in.Plan.RepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next repo number: %w", err)
	}

	// 先创建扫描任务（ExecutionID 暂为 0，后续更新）
	task := &scanStore.ScanTask{
		PlanID:     uint64(in.Plan.ID),
		RepoNumber: repoNumber,
		PlanNumber: planNumber,
		ScanType:   constant.ScanType(in.Plan.ScanType),
		Status:     constant.ScanTaskStatusInProgress,
		CreatedBy:  in.Hook.Sender,
	}
	if in.Trigger == nil {
		task.TriggerType = constant.ScanPlanTriggerTypeManual
	} else {
		task.TriggerType = in.Trigger.Type
		if in.Trigger.ID != 0 {
			task.TriggerID = uint64(in.Trigger.ID)
		}
	}

	created, err := s.taskStore.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan task: %w", err)
	}

	// 注入扫描参数到 Hook.Params，随流水线传递到容器环境变量
	var triggerID int64
	if in.Trigger != nil {
		triggerID = in.Trigger.ID
	}
	in.Hook.Params["INNER_PIPELINE"] = "true"
	in.Hook.Params["SCAN_PLAN_ID"] = fmt.Sprintf("%d", in.Plan.ID)
	in.Hook.Params["SCAN_TASK_ID"] = fmt.Sprintf("%d", created.ID)
	in.Hook.Params["SCAN_TRIGGER_ID"] = fmt.Sprintf("%d", triggerID)
	in.Hook.Params["SCAN_REPO_ID"] = fmt.Sprintf("%d", in.Repo.ID)
	// 注入 git clone 需要的参数
	in.Hook.Params["gitURL"] = in.Repo.GitURL
	in.Hook.Params["gitRef"] = in.Hook.Ref
	in.Hook.Params["PLUGIN_WORKSPACE"] = in.Repo.Identifier // 供流水线中插件使用，表示代码 clone 的目录

	// 获取或创建代码扫描流水线
	pipeline, err := s.getOrCreatePipeline(ctx, in.Plan.RepoID, in)
	if err != nil {
		return nil, err
	}

	// 触发流水线执行
	execution, err := s.triggerSvc.Trigger(ctx, uint(pipeline.ID), in.Hook)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger pipeline: %w", err)
	}

	// 用 execution.ID 回填扫描任务
	if err := s.taskStore.Update(ctx, &scanStore.ScanTask{
		ID:          uint64(created.ID),
		ExecutionID: uint64(execution.ID),
	}); err != nil {
		return nil, fmt.Errorf("failed to update scan task execution id: %w", err)
	}
	created.ExecutionID = int64(execution.ID)

	return created, nil
}

// CreateRunnerTask 创建流水线 runner 回调的扫描任务（非内嵌流水线场景）。
// codescan 插件以普通插件方式运行于流水线（gitfox 未注入 SCAN_TASK_ID，INNER_PIPELINE 非
// "true"）时，由插件在扫描前调用本方法创建任务记录、换取上报报告所需的 taskID。
// 提交范围不在这里解析：已改由插件的 SCAN_METHOD + SCAN_REF_RANGE 声明。
// 与 Create 不同：流水线已在执行中，这里不创建/触发代码扫描流水线，只按 body 透传的
// executionID/triggerType 落任务记录（triggerID 不落库，见函数内注释）。
func (s *ScanTaskService) CreateRunnerTask(ctx context.Context, in *scanModel.ScanTaskRunnerCreateReq) (*scanModel.ScanTaskRunnerCreateResp, error) {
	// 根据 body 中的 repoID/planID 查询 repo 和 plan，并校验 plan 归属
	repo, err := s.repoStore.GetByID(ctx, uint(in.RepoID))
	if err != nil {
		return nil, fmt.Errorf("failed to get repo by ID %d: %w", in.RepoID, err)
	}
	plan, err := s.planStore.GetByID(ctx, uint(in.PlanID))
	if err != nil {
		return nil, fmt.Errorf("failed to get scan plan by ID %d: %w", in.PlanID, err)
	}
	if plan.RepoID != int64(repo.ID) {
		return nil, fmt.Errorf("scan plan %d does not belong to repo %d", in.PlanID, in.RepoID)
	}

	// 计算序号
	planNumber, err := s.taskStore.NextPlanNumber(ctx, uint64(plan.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get next plan number: %w", err)
	}
	repoNumber, err := s.taskStore.NextRepoNumber(ctx, plan.RepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next repo number: %w", err)
	}

	// 创建扫描任务记录，ExecutionID 直接取流水线执行 ID
	task := &scanStore.ScanTask{
		PlanID:      uint64(plan.ID),
		ExecutionID: uint64(in.ExecutionID),
		RepoNumber:  repoNumber,
		PlanNumber:  planNumber,
		ScanType:    constant.ScanType(plan.ScanType),
		Status:      constant.ScanTaskStatusInProgress,
	}
	// 插件透传的是 GITFOX_BUILD_TRIGGER（TriggerAction 原值），归一为任务表记录的 manual/cron/action
	task.TriggerType = normalizeScanTriggerType(in.TriggerType)
	// 不落 triggerID：这类任务由普通流水线触发，其触发器与 scan 触发器不同表，
	// triggerID 只由内置 scan 流水线（Create/Retry）写入

	created, err := s.taskStore.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan task: %w", err)
	}
	return &scanModel.ScanTaskRunnerCreateResp{TaskID: int64(created.ID)}, nil
}

// listLatestCommits 取 branch 最新两笔提交（倒序：[最新, 次新]）。
// 包级函数：扫描任务（ScanTaskService）与 AI 评审（AIReviewService）共用。
func listLatestCommits(ctx context.Context, lister CommitLister, repoUID, branch string) ([]string, error) {
	shas, err := lister.ListCommitSHAs(ctx, &git.ListCommitsParams{
		ReadParams: git.ReadParams{RepoUID: repoUID},
		GitREF:     branch,
		Page:       1,
		Limit:      2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list commit SHAs for branch %s: %w", branch, err)
	}
	return shas, nil
}

// ResolveCommitRange 按扫描计划的 scanType 解析 branch 的提交范围，供内嵌扫描流水线
// 的两条路径共用：定时触发（cron handler 的 scanTriggerHandler）与任务重试（Retry）。
// 增量取最新两笔（次新为 before、最新为 after，插件按 git diff 过滤报告行），
// 全量取同一笔（before = after = 最新，插件走全量扫描）。scanType 为空或未知按全量。
//
// 普通流水线（codescan 插件的 create-by-runner）不再走这里：提交范围由插件的
// SCAN_METHOD + SCAN_REF_RANGE 声明。
func (s *ScanTaskService) ResolveCommitRange(ctx context.Context, repoUID string, scanType constant.ScanType, branch string) (commitBefore, commitAfter string, err error) {
	if branch == "" {
		return "", "", fmt.Errorf("扫描分支为空，无法确定扫描范围")
	}
	shas, err := listLatestCommits(ctx, s.commitLister, repoUID, branch)
	if err != nil {
		return "", "", err
	}
	return buildCommitRange(scanType, branch, shas)
}

// buildCommitRange 是 ResolveCommitRange 的纯函数核心，便于单测。
// 增量（scanType=incremental）→ 次新/最新；其余（full、scanType 为空或未知）→ 最新/最新。
func buildCommitRange(scanType constant.ScanType, branch string, shas []string) (commitBefore, commitAfter string, err error) {
	if scanType == constant.ScanTypeIncremental {
		return buildLatestTwoCommitRange(branch, shas)
	}
	if len(shas) == 0 {
		return "", "", fmt.Errorf("no commits found on branch %s", branch)
	}
	return shas[0], shas[0], nil
}

// buildLatestTwoCommitRange 是增量口径的纯函数核心：最新为 after、次新为 before；
// 分支只有一笔提交时两者相同（增量无从谈起，插件按全量处理）。
// shas 为 ListCommitSHAs(ref, page=1, limit=2) 的结果，按时间倒序：[最新, 次新]。
func buildLatestTwoCommitRange(branch string, shas []string) (commitBefore, commitAfter string, err error) {
	if len(shas) == 0 {
		return "", "", fmt.Errorf("no commits found on branch %s", branch)
	}
	if len(shas) == 1 {
		return shas[0], shas[0], nil
	}
	return shas[1], shas[0], nil
}

// normalizeScanTriggerType 把插件透传的 GITFOX_BUILD_TRIGGER 归一为扫描任务表
// 记录的 manual/cron/action。该值取自 execution.Trigger（即 Hook.Action，为空时
// triggerer 兜底为 "manual"），可能是 branch_updated/pullreq_created 等 TriggerAction
// 原值，也可能是非枚举的 "manual"。
func normalizeScanTriggerType(triggerType string) constant.ScanPlanTriggerType {
	switch triggerType {
	case "", string(constant.TriggerEventManual):
		return constant.ScanPlanTriggerTypeManual
	case string(constant.TriggerActionCron):
		return constant.ScanPlanTriggerTypeCron
	default:
		return constant.ScanPlanTriggerTypeAction
	}
}

func (s *ScanTaskService) CreateBranchHook(ctx context.Context, repo *repoModel.Repo, branch string, sha, beforeSha string, principal *authModel.Session) (*triggerer.Hook, error) {
	hook := &triggerer.Hook{
		Commit:       sha,
		CommitBefore: beforeSha,
		Trigger:      principal.User.Account,
		TriggeredBy:  principal.User.ID,
		Action:       "",
		Parent:       int64(repo.SpaceID),
		Ref:          branch,
		Sender:       principal.User.Account,
		Params:       map[string]string{},
	}
	return hook, nil
}

// Retry 重试失败的扫描任务
func (s *ScanTaskService) Retry(ctx context.Context, session *authModel.Session, taskID uint, repo *repoModel.Repo) (*scanModel.ScanTask, error) {
	// 获取任务详情
	taskInfo, err := s.taskStore.GetTaskInfo(ctx, taskID)
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to get scan task info: %w", err), i18n.ScanTaskRetryError)
	}

	// 只有失败的任务才能重试
	if taskInfo.Status != constant.ScanTaskStatusFailed {
		return nil, errors.WithI18nKey(errors.New("only failed tasks can be retried"), i18n.ScanTaskRetryNotFailed)
	}

	// 获取扫描计划
	plan, err := s.planStore.GetByID(ctx, uint(taskInfo.PlanID))
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to get scan plan: %w", err), i18n.ScanTaskRetryError)
	}

	// 解析重试的扫描范围：复用原执行的分支（手动扫描为用户当时选择的分支，见
	// ExecutionRef 的来源 ops_pipeline_executions.ref），按当时扫描计划的 scanType
	// 取该分支的提交范围。ref 已不存在（分支被删、PR ref 被清理）时直接失败——今天
	// 这种重试也会失败在 git clone 阶段，提前失败文案更清楚
	commitBefore, commitAfter, err := s.ResolveCommitRange(ctx, repo.GitUID, constant.ScanType(plan.ScanType), taskInfo.ExecutionRef)
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to resolve commit range for retry: %w", err), i18n.ScanTaskRetryError)
	}

	// 构建重试 Hook：与手动扫描共用 CreateBranchHook（其形参为 sha=最新、beforeSha=次新），
	// 保证提交范围口径一致
	hook, err := s.CreateBranchHook(ctx, repo, taskInfo.ExecutionRef, commitAfter, commitBefore, session)
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to create retry hook: %w", err), i18n.ScanTaskRetryError)
	}
	hook.Timestamp = time.Now().UnixMilli()

	// 注入扫描参数到 Hook.Params
	var triggerID int64
	if taskInfo.Trigger != nil {
		triggerID = taskInfo.Trigger.ID
	}

	hook.Params["INNER_PIPELINE"] = "true"
	hook.Params["SCAN_PLAN_ID"] = fmt.Sprintf("%d", taskInfo.PlanID)
	hook.Params["SCAN_TASK_ID"] = fmt.Sprintf("%d", taskID)
	hook.Params["SCAN_TRIGGER_ID"] = fmt.Sprintf("%d", triggerID)
	hook.Params["SCAN_REPO_ID"] = fmt.Sprintf("%d", taskInfo.RepoID)
	// 注入 git clone 需要的参数
	hook.Params["gitURL"] = repo.GitURL
	hook.Params["gitRef"] = hook.Ref
	hook.Params["PLUGIN_WORKSPACE"] = repo.Identifier // 供流水线中插件使用，表示代码 clone 的目录
	// 注入 gitfoxscan 调用接口时需要的参数
	hook.Params["GITFOX_API_ENDPOINT"] = s.config.URL.Base + "/api/v2"
	hook.Params["PLUGIN_TOKEN"] = s.config.PipelineAuth.Password

	// 获取或创建流水线
	input := &ScanTaskInput{
		Plan: plan,
		Hook: hook,
	}
	pipeline, err := s.getOrCreatePipeline(ctx, plan.RepoID, input)
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to get pipeline: %w", err), i18n.ScanTaskRetryError)
	}

	// 触发新执行
	execution, err := s.triggerSvc.Trigger(ctx, uint(pipeline.ID), hook)
	if err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to trigger retry: %w", err), i18n.ExecutionTriggerError)
	}

	// 更新原任务关联新执行
	task := &scanStore.ScanTask{
		ID:          uint64(taskID),
		ExecutionID: uint64(execution.ID),
		Status:      constant.ScanTaskStatusInProgress,
	}
	if err := s.taskStore.Update(ctx, task); err != nil {
		return nil, errors.WithI18nKey(fmt.Errorf("failed to update scan task: %w", err), i18n.ScanTaskRetryError)
	}

	return s.taskStore.GetByID(ctx, taskID)
}

// GetTaskInfo 获取扫描任务详情
func (s *ScanTaskService) GetTaskInfo(ctx context.Context, taskID uint) (*scanModel.ScanTaskInfo, error) {
	return s.taskStore.GetTaskInfo(ctx, taskID)
}

// ListTaskInfos 查询扫描任务详情列表
func (s *ScanTaskService) ListTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (*database.PageResult[*scanModel.ScanTaskInfo], error) {
	return s.taskStore.ListTaskInfos(ctx, filter)
}

// ListTaskIssues 查询指定任务关联的 issue 列表
func (s *ScanTaskService) ListTaskIssues(ctx context.Context, taskID uint, filter scanModel.ScanTaskIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error) {
	return s.issueStore.ListByTaskID(ctx, taskID, filter)
}

// CountTaskInfos 统计扫描任务数量
func (s *ScanTaskService) CountTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (int64, error) {
	return s.taskStore.CountTaskInfos(ctx, filter)
}

// Delete 删除扫描任务
func (s *ScanTaskService) Delete(ctx context.Context, id uint) error {
	return s.taskStore.Delete(ctx, id)
}

// SetupCron 设置定时扫描任务
func (s *ScanTaskService) SetupCron(ctx context.Context, trigger *scanModel.ScanTrigger) error {
	if s.cronScheduler == nil {
		return fmt.Errorf("cron scheduler not configured")
	}
	identifier := &JobCodeScanIdentifier{
		RepoID:    trigger.RepoID,
		PlanID:    trigger.PlanID,
		TriggerID: trigger.ID,
	}
	dataByte, _ := json.Marshal(identifier)
	if err := s.cronScheduler.AddRecurringWithData(ctx, identifier.String(), JobTypeCodeScan, string(dataByte), trigger.Cron, JobMaxDurationCodeScan); err != nil {
		return fmt.Errorf("failed to setup cron: %w", err)
	}
	return nil
}

// RecycleCron 取消定时扫描任务
func (s *ScanTaskService) RecycleCron(ctx context.Context, trigger *scanModel.ScanTrigger) error {
	if s.cronScheduler == nil {
		return fmt.Errorf("cron scheduler not configured")
	}
	identifier := &JobCodeScanIdentifier{
		RepoID:    trigger.RepoID,
		PlanID:    trigger.PlanID,
		TriggerID: trigger.ID,
	}
	if err := s.cronScheduler.RemoveRecurring(ctx, identifier.String()); err != nil {
		return fmt.Errorf("failed to recycle cron: %w", err)
	}
	return nil
}

// ==================== 内部方法 ====================

// getOrCreatePipeline 获取或创建代码扫描流水线
func (s *ScanTaskService) getOrCreatePipeline(ctx context.Context, repoID int64, in *ScanTaskInput) (*pipelineModel.Pipeline, error) {
	pipeline, err := s.pipelineStore.GetPipelineByRepoIDAndName(ctx, repoID, _codeScanPipelineName)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		// 这里逻辑需要变更，如果没有找到流水线，除了创建流水线，还需要创建流水线对应的 content
		pipeline, err = s.createPipeline(ctx, repoID, in)
		if err != nil {
			return nil, err
		}
	} else {
		// 如果流水线已存在，更新流水线对应的 content（覆盖原有内容）
		if err := s.updatePipelineContent(ctx, pipeline.ID, in); err != nil {
			return nil, err
		}
	}

	return pipeline, nil
}

// createPipeline 创建代码扫描流水线
func (s *ScanTaskService) createPipeline(ctx context.Context, repoID int64, in *ScanTaskInput) (*pipelineModel.Pipeline, error) {
	if s.containsLangCsharp(ctx, in.Plan.ID) {
		return nil, errors.WithI18nKey(nil, i18n.PipelineCSharpNotSupportError)
	}

	// 创建流水线
	pipelineType := "repo"
	if repoID == 0 {
		pipelineType = "space"
	}
	p := &pipelineModel.Pipeline{
		Name:          _codeScanPipelineName,
		Scope:         pipelineType,
		RepoID:        repoID,
		Desc:          _codeScanDescription,
		YamlPath:      pipelineModel.PipelineConfigPathInnerPrefix + _codeScanLabel + ".yaml",
		DefaultBranch: in.Hook.Ref,
		CreatedBy:     in.Hook.Sender,
	}

	// 使用公共方法创建内置流水线
	return s.pipelineSvc.CreateEmbeddedPipeline(ctx, _codeScanLabel, p)
}

// updatePipelineContent 更新代码扫描流水线内容
func (s *ScanTaskService) updatePipelineContent(ctx context.Context, pipelineID int64, in *ScanTaskInput) error {
	if s.containsLangCsharp(ctx, in.Plan.ID) {
		return errors.WithI18nKey(nil, i18n.PipelineCSharpNotSupportError)
	}

	// 使用公共方法更新内置流水线内容
	return s.pipelineSvc.UpdateEmbeddedPipelineContent(ctx, _codeScanLabel, pipelineID)
}

// containsLangCsharp 检查扫描计划关联的方案是否包含 C# 语言
func (s *ScanTaskService) containsLangCsharp(ctx context.Context, planID int64) bool {
	solutionIDs, err := s.planStore.GetSolutionIDList(ctx, uint(planID))
	if err != nil || len(solutionIDs) == 0 {
		return false
	}
	for _, slnID := range solutionIDs {
		rulesets, err := s.solutionStore.ListRulesets(ctx, uint(slnID))
		if err != nil {
			continue
		}
		for _, rs := range rulesets {
			if rs.Lang == constant.ScanLangCsharpName {
				return true
			}
		}
	}
	return false
}
