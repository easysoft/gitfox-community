package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
)

// cronTriggerHandler implements pkgJobs.Handler for pipeline cron triggers.
type cronTriggerHandler struct {
	triggerer   triggerer.Triggerer
	pipelineSvc pipelineSvc.Interface
}

type cronJobInput struct {
	PipelineID int64 `json:"pipelineID"`
}


func (h *cronTriggerHandler) Handle(ctx context.Context, input string, _ pkgJobs.ProgressReporter) (string, error) {
	var in cronJobInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		return "", fmt.Errorf("failed to parse cron job input: %w", err)
	}

	if in.PipelineID <= 0 {
		return "", fmt.Errorf("invalid pipelineID %d in cron job input", in.PipelineID)
	}

	_, err := h.pipelineSvc.GetPipelineByID(ctx, in.PipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline %d: %w", in.PipelineID, err)
	}

	// 从 ops_pipeline_content 解析 variables，取 key=gitRef 的分支值。
	// 优先取 value，value 为空则取 defaultValue，两者均为空则报错。
	gitRef, err := h.resolveGitRef(ctx, in.PipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve gitRef for pipeline %d: %w", in.PipelineID, err)
	}

	hook := &triggerer.Hook{
		Action:  constant.TriggerActionCron,
		Ref:     gitRef,
		Trigger: "system",
	}

	//nolint:gosec // PipelineID is validated > 0 above, safe to convert to uint.
	_, err = h.triggerer.Trigger(ctx, uint(in.PipelineID), hook)
	if err != nil {
		return "", fmt.Errorf("failed to trigger pipeline %d: %w", in.PipelineID, err)
	}

	return "", nil
}

// resolveGitRef 从流水线 variables 中查找 key=gitRef 的值并返回。
// 使用 GetPipelineVariables：runtime=true 时取 defaultValue（定时任务无人提供运行时值），
// runtime=false 时取 value（固定值）。两者皆空则返回错误。
func (h *cronTriggerHandler) resolveGitRef(ctx context.Context, pipelineID int64) (string, error) {
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

// RegisterPipelineCronHandler registers the cron trigger handler with the job
// scheduler. Must be called before Scheduler.Run().
func RegisterPipelineCronHandler(
	scheduler *jobsSvc.Scheduler,
	t triggerer.Triggerer,
	ps pipelineSvc.Interface,
) error {
	handler := &cronTriggerHandler{
		triggerer:   t,
		pipelineSvc: ps,
	}
	return scheduler.RegisterHandler(pipelineSvc.JobTypePipelineCron, handler)
}
