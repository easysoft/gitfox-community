package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	principalModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/httpclient"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

type pipelineTriggerHandler struct {
	triggerer    triggerer.Triggerer
	pipelineSvc  pipelineSvc.Interface
	repoSvc      repoSvc.Interface
	devopsConfig config.DevopsConfig
	tokenStore   token.Interface
}

type pipelineJobInput struct {
	PipelineID int64 `json:"pipelineID"`
}

func (h *pipelineTriggerHandler) Handle(ctx context.Context, input string, _ pkgJobs.ProgressReporter) (string, error) {
	var in pipelineJobInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		return "", fmt.Errorf("failed to parse cron job input: %w", err)
	}
	if in.PipelineID <= 0 {
		return "", fmt.Errorf("invalid pipelineID %d in cron job input", in.PipelineID)
	}
	pipeline, err := h.pipelineSvc.GetPipelineByID(ctx, in.PipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline %d: %w", in.PipelineID, err)
	}

	switch pipeline.Engine {
	case constant.PipelineEngineGitLab, constant.PipelineEngineJenkins:
		return h.triggerExternalPipeline(ctx, pipeline)
	default:
		return h.triggerPipeline(ctx, pipeline)
	}
}

// triggerExternalPipeline 调用 DevOps 接口触发外部流水线
func (h *pipelineTriggerHandler) triggerExternalPipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) (string, error) {
	entry, err := h.tokenStore.GetDevopsToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get devops token for external pipeline trigger: %w", err)
	}

	url := fmt.Sprintf("%s/pipelinetrigger?id=%d", h.devopsConfig.URL, pipeline.ID)
	headers := map[string]string{
		"Authorization": entry.Key,
	}

	log := logger.GetLogger(ctx)
	log.Infow("Triggering external pipeline", "pipelineID", pipeline.ID, "url", url)

	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, nil); err != nil {
		return "", fmt.Errorf("failed to trigger external pipeline %d: %w", pipeline.ID, err)
	}
	return "", nil
}

// triggerPipeline 通过内部 triggerer 触发流水线
func (h *pipelineTriggerHandler) triggerPipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) (string, error) {
	// 流水线变量整体作为运行参数注入（与事件触发保持一致）
	vars, err := h.pipelineSvc.GetPipelineVariables(ctx, pipeline.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline variables for pipeline %d: %w", pipeline.ID, err)
	}
	// gitRef 未配置时为空串，不报错
	gitRef := vars["gitRef"]

	params := make(map[string]string, len(vars)+2)
	for k, v := range vars {
		params[k] = v
	}
	params["gitRef"] = gitRef

	hook := &triggerer.Hook{
		Action:  constant.TriggerActionCron,
		Ref:     gitRef,
		Trigger: "system",
		Params:  params,
	}

	if pipeline.RepoID != 0 {
		session := &authModel.Session{
			Principal: &principalModel.Principal{
				App:         "system",
				DisplayName: "system",
				Admin:       true,
			},
			User: &principalModel.User{
				Account: "system",
			},
		}
		repo, err := h.repoSvc.GetByID(ctx, session, uint(pipeline.RepoID))
		if err != nil {
			return "", fmt.Errorf("failed to get repo %d for pipeline %d: %w", pipeline.RepoID, pipeline.ID, err)
		}
		hook.Params["gitURL"] = repo.GitURL
	}

	if _, err = h.triggerer.Trigger(ctx, uint(pipeline.ID), hook); err != nil {
		return "", fmt.Errorf("failed to trigger pipeline %d: %w", pipeline.ID, err)
	}
	return "", nil
}

type scanTriggerHandler struct {
	taskSvc      scanSvc.IScanTask
	planSvc      scanSvc.IScanPlan
	triggerStore scanStore.IScanTrigger
	repoSvc      repoSvc.Interface
}

type scanJobInput struct {
	RepoID    int64 `json:"repoID"`
	PlanID    int64 `json:"planID"`
	TriggerID int64 `json:"triggerID"`
}

func (h *scanTriggerHandler) Handle(ctx context.Context, input string, _ pkgJobs.ProgressReporter) (string, error) {
	var in scanJobInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		return "", err
	}

	session := &authModel.Session{
		Principal: &principalModel.Principal{
			App:         "system",
			DisplayName: "system",
			Admin:       true,
		},
		User: &principalModel.User{
			Account: "system",
		},
	}

	repo, err := h.repoSvc.GetByID(ctx, session, uint(in.RepoID))
	if err != nil {
		return "", err
	}

	plan, err := h.planSvc.GetByID(ctx, uint(in.RepoID), uint(in.PlanID))
	if err != nil {
		return "", err
	}

	trigger, err := h.triggerStore.GetByID(ctx, uint(in.TriggerID))
	if err != nil {
		return "", err
	}

	if trigger.Type != scanModel.ScanTriggerTypeCron {
		return "", fmt.Errorf("invalid scan trigger type: %s", trigger.Type)
	}

	if trigger.CronBranch == "" {
		return "", fmt.Errorf("invalid scan trigger cron branch: %s", trigger.CronBranch)
	}

	// 按扫描计划的 scanType 解析提交范围：增量取最新两笔、全量取同一笔
	sha, beforeSha, err := h.taskSvc.ResolveCommitRange(ctx, repo.GitUID, constant.ScanType(plan.ScanType), trigger.CronBranch)
	if err != nil {
		return "", err
	}

	hook, err := h.taskSvc.CreateBranchHook(ctx, repo, trigger.CronBranch, sha, beforeSha, session)
	if err != nil {
		return "", err
	}
	hook.Trigger = "cron"
	hook.Action = constant.TriggerActionCron

	if _, err = h.taskSvc.Create(ctx, &scanSvc.ScanTaskInput{
		Plan: plan,
		Repo: repo,
		Trigger: &scanModel.ScanTaskTrigger{
			ID:   trigger.ID,
			Name: trigger.Name,
			Type: constant.ScanPlanTriggerTypeCron,
		},
		Hook: hook,
	}); err != nil {
		return "", err
	}
	return "", nil
}

func RegisterPipelineHandler(scheduler *jobsSvc.Scheduler, t triggerer.Triggerer, svc pipelineSvc.Interface, rs repoSvc.Interface) error {
	handler := &pipelineTriggerHandler{triggerer: t, pipelineSvc: svc, repoSvc: rs}
	return scheduler.RegisterHandler(pipelineSvc.JobTypePipelineCron, handler)
}

func RegisterGitlabPipelineHandler(scheduler *jobsSvc.Scheduler, t triggerer.Triggerer, svc pipelineSvc.Interface, rs repoSvc.Interface, devopsConfig config.DevopsConfig, tokenStore token.Interface) error {
	handler := &pipelineTriggerHandler{triggerer: t, pipelineSvc: svc, repoSvc: rs, devopsConfig: devopsConfig, tokenStore: tokenStore}
	return scheduler.RegisterHandler(pipelineSvc.JobTypeGitLabPipelineCron, handler)
}

func RegisterJenkinsPipelineHandler(scheduler *jobsSvc.Scheduler, t triggerer.Triggerer, svc pipelineSvc.Interface, rs repoSvc.Interface, devopsConfig config.DevopsConfig, tokenStore token.Interface) error {
	handler := &pipelineTriggerHandler{triggerer: t, pipelineSvc: svc, repoSvc: rs, devopsConfig: devopsConfig, tokenStore: tokenStore}
	return scheduler.RegisterHandler(pipelineSvc.JobTypeJenkinsPipelineCron, handler)
}

// 定时触发外部流水线同步
type pipelineSyncHandler struct {
	devopsConfig config.DevopsConfig
	tokenStore   token.Interface
}

func (h *pipelineSyncHandler) Handle(ctx context.Context, input string, _ pkgJobs.ProgressReporter) (string, error) {

	entry, err := h.tokenStore.GetDevopsToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get devops token for external pipeline sync: %w", err)
	}

	url := fmt.Sprintf("%s/syncexternalpipeline", h.devopsConfig.URL)
	headers := map[string]string{
		"Authorization": entry.Key,
	}

	log := logger.GetLogger(ctx)
	log.Infow("sync external pipeline", "url", url)

	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, nil); err != nil {
		return "", fmt.Errorf("failed to sync external pipeline: %w", err)
	}
	return "", nil
}

func RegisterPipelineSyncHandler(scheduler *jobsSvc.Scheduler, devopsConfig config.DevopsConfig, tokenStore token.Interface) error {
	handler := &pipelineSyncHandler{devopsConfig: devopsConfig, tokenStore: tokenStore}
	return scheduler.RegisterHandler(pipelineSvc.JobPipelineSyncCron, handler)
}

// RegisterArtifactGCHandler 注册制品 GC 的 cron handler。
// 仅注册通用 JobTypeArtifactGC：自动触发清理全部格式（entity 留空），
// 用户主动调用 /artifacts/garbage[/:entity] 时按指定 entity 清理。
func RegisterArtifactGCHandler(scheduler *jobsSvc.Scheduler, svc artifactSvc.Interface) error {
	return scheduler.RegisterHandler(artifactSvc.JobTypeArtifactGC, artifactGCHandler{svc: svc})
}

type artifactGCHandler struct {
	svc artifactSvc.Interface
}

// artifactGCInput 是制品 GC 任务的入参。
//   - format 为空 → 清理全部格式（自动触发场景）
//   - format 指定具体 format → 仅清理该格式（用户主动 API 触发场景）
type artifactGCInput struct {
	Format string `json:"format,omitempty"`
}

func (h artifactGCHandler) Handle(ctx context.Context, input string, _ pkgJobs.ProgressReporter) (string, error) {
	var in artifactGCInput
	if input != "" {
		if err := json.Unmarshal([]byte(input), &in); err != nil {
			return "", fmt.Errorf("failed to parse artifact gc input: %w", err)
		}
	}
	_, err := h.svc.GarbageCollect(ctx, in.Format)
	return "", err
}

func RegisterScanHandler(
	scheduler *jobsSvc.Scheduler,
	taskSvc scanSvc.IScanTask,
	planSvc scanSvc.IScanPlan,
	triggerStore scanStore.IScanTrigger,
	rs repoSvc.Interface,
) error {
	handler := &scanTriggerHandler{taskSvc: taskSvc, planSvc: planSvc, triggerStore: triggerStore, repoSvc: rs}
	return scheduler.RegisterHandler(scanSvc.JobTypeScanCron, handler)
}
