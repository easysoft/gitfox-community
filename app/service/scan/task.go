package scan

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
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
	config         *config.Config
	taskStore      scanStore.IScanTask
	planStore      scanStore.IScanPlan
	solutionStore  scanStore.ISolution
	issueStore     scanStore.IScanIssue
	conditionStore scanStore.IScanCondition
	ruleSetStore   scanStore.IRuleSet
	ruleStore      scanStore.IRule
	repoStore      repoStore.Interface
	pipelineStore  pipelineStore.Interface
	triggerSvc     triggerer.Triggerer
	cronScheduler  CronScheduler
	tx             dbtx.TxRunner
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
	triggerSvc triggerer.Triggerer,
	cronScheduler CronScheduler,
	tx dbtx.TxRunner,
) *ScanTaskService {
	return &ScanTaskService{
		config:         config,
		taskStore:      taskStore,
		planStore:      planStore,
		solutionStore:  solutionStore,
		issueStore:     issueStore,
		conditionStore: conditionStore,
		ruleSetStore:   ruleSetStore,
		ruleStore:      ruleStore,
		repoStore:      repoStore,
		pipelineStore:  pipelineStore,
		triggerSvc:     triggerSvc,
		cronScheduler:  cronScheduler,
		tx:             tx,
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
			triggerID := uint64(in.Trigger.ID)
			task.TriggerID = &triggerID
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
	in.Hook.Params["SCAN_PLAN_ID"] = fmt.Sprintf("%d", in.Plan.ID)
	in.Hook.Params["SCAN_TASK_ID"] = fmt.Sprintf("%d", created.ID)
	in.Hook.Params["SCAN_TRIGGER_ID"] = fmt.Sprintf("%d", triggerID)
	in.Hook.Params["SCAN_REPO_ID"] = fmt.Sprintf("%d", in.Repo.ID)
	// 注入 git clone 需要的参数
	in.Hook.Params["gitURL"] = in.Repo.GitURL
	in.Hook.Params["gitRef"] = in.Hook.Ref
	in.Hook.Params["PLUGIN_WORKSPACE"] = in.Repo.Identifier // 供流水线中插件使用，表示代码 clone 的目录
	// 注入 gitfoxscan 调用接口时需要的参数
	in.Hook.Params["GITFOX_API_ENDPOINT"] = s.config.URL.Base + "/api/v2"
	in.Hook.Params["PLUGIN_TOKEN"] = s.config.PipelineAuth.Password

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

func (s *ScanTaskService) CreateBranchHook(ctx context.Context, repo *repoModel.Repo, branch string, sha string, principal *authModel.Session) (*triggerer.Hook, error) {
	hook := &triggerer.Hook{
		Commit:      sha,
		Trigger:     principal.User.Account,
		TriggeredBy: principal.User.ID,
		Action:      "",
		Parent:      int64(repo.SpaceID),
		Ref:         branch,
		Sender:      principal.User.Account,
		Params:      map[string]string{},
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

	// 构建重试 Hook
	hook := &triggerer.Hook{
		Trigger:     session.User.Account,
		TriggeredBy: session.User.ID,
		Ref:         taskInfo.ExecutionRef,
		Sender:      session.User.Account,
		Params:      map[string]string{},
		Timestamp:   time.Now().UnixMilli(),
	}

	// 注入扫描参数到 Hook.Params
	var triggerID int64
	if taskInfo.Trigger != nil {
		triggerID = taskInfo.Trigger.ID
	}

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
	yamlFile := "pipeline/" + _codeScanLabel + ".yaml"

	// 读取 embed 的流水线配置
	data, err := resources.EmbeddedPipelines.ReadFile(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded pipeline config %s: %w", yamlFile, err)
	}

	// 创建流水线
	pipelineType := "repo"
	if repoID == 0 {
		pipelineType = "space"
	}
	p := &pipelineModel.Pipeline{
		Name:          _codeScanPipelineName,
		Type:          pipelineType,
		RepoID:        repoID,
		Desc:          _codeScanDescription,
		YamlPath:      pipelineModel.PipelineConfigPathInnerPrefix + _codeScanLabel + ".yaml",
		DefaultBranch: in.Hook.Ref,
		CreatedBy:     in.Hook.Sender,
	}
	pipeline, err := s.pipelineStore.CreatePipeline(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// 创建流水线内容（base64 编码后存入）
	content := &pipelineModel.PipelineContent{
		PipelineID: pipeline.ID,
		Data:       []byte(base64.StdEncoding.EncodeToString(data)),
		CreatedBy:  in.Hook.Sender,
	}
	if _, err = s.pipelineStore.CreatePipelineContent(ctx, content); err != nil {
		return nil, fmt.Errorf("failed to create pipeline content: %w", err)
	}

	return pipeline, nil
}

// updatePipelineContent 更新代码扫描流水线内容
func (s *ScanTaskService) updatePipelineContent(ctx context.Context, pipelineID int64, in *ScanTaskInput) error {
	if s.containsLangCsharp(ctx, in.Plan.ID) {
		return errors.WithI18nKey(nil, i18n.PipelineCSharpNotSupportError)
	}
	yamlFile := "pipeline/" + _codeScanLabel + ".yaml"

	// 读取 embed 的流水线配置
	data, err := resources.EmbeddedPipelines.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read embedded pipeline config %s: %w", yamlFile, err)
	}
	// 创建流水线内容（base64 编码后存入）
	content := &pipelineModel.PipelineContent{
		PipelineID: pipelineID,
		Data:       []byte(base64.StdEncoding.EncodeToString(data)),
	}
	return s.pipelineStore.UpdatePipelineContent(ctx, content)
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
