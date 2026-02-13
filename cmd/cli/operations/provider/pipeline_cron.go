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

	pipeline, err := h.pipelineSvc.GetPipelineByID(ctx, in.PipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline %d: %w", in.PipelineID, err)
	}

	hook := &triggerer.Hook{
		Action:  constant.TriggerActionCron,
		Ref:     pipeline.DefaultBranch,
		Trigger: "system",
	}

	//nolint:gosec // PipelineID is validated > 0 above, safe to convert to uint.
	_, err = h.triggerer.Trigger(ctx, uint(in.PipelineID), hook)
	if err != nil {
		return "", fmt.Errorf("failed to trigger pipeline %d: %w", in.PipelineID, err)
	}

	return "", nil
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
