package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	principalModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
)

type pipelineTriggerHandler struct {
	triggerer   triggerer.Triggerer
	pipelineSvc pipelineSvc.Interface
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
	if _, err := h.pipelineSvc.GetPipelineByID(ctx, in.PipelineID); err != nil {
		return "", fmt.Errorf("failed to get pipeline %d: %w", in.PipelineID, err)
	}

	gitRef, err := h.resolveGitRef(ctx, in.PipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve gitRef for pipeline %d: %w", in.PipelineID, err)
	}

	hook := &triggerer.Hook{
		Action:  constant.TriggerActionCron,
		Ref:     gitRef,
		Trigger: "system",
		Params:  map[string]string{"gitRef": gitRef},
	}

	if _, err = h.triggerer.Trigger(ctx, uint(in.PipelineID), hook); err != nil {
		return "", fmt.Errorf("failed to trigger pipeline %d: %w", in.PipelineID, err)
	}
	return "", nil
}

func (h *pipelineTriggerHandler) resolveGitRef(ctx context.Context, pipelineID int64) (string, error) {
	vars, err := h.pipelineSvc.GetPipelineVariables(ctx, pipelineID)
	if err != nil {
		return "", fmt.Errorf("pipeline %d: failed to get pipeline variables: %w", pipelineID, err)
	}
	ref, ok := vars["gitRef"]
	if !ok {
		return "", fmt.Errorf("pipeline %d: gitRef variable not found in variables, please configure it", pipelineID)
	}
	if ref == "" {
		return "", fmt.Errorf("pipeline %d: gitRef variable is not set", pipelineID)
	}
	return ref, nil
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

	hook, err := h.taskSvc.CreateBranchHook(ctx, repo, trigger.CronBranch, "", session)
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

func RegisterPipelineHandler(scheduler *jobsSvc.Scheduler, t triggerer.Triggerer, svc pipelineSvc.Interface) error {
	handler := &pipelineTriggerHandler{triggerer: t, pipelineSvc: svc}
	return scheduler.RegisterHandler(pipelineSvc.JobTypePipelineCron, handler)
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
