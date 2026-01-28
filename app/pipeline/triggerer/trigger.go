package triggerer

import (
	"context"
	"runtime/debug"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

var _ Triggerer = (*triggerer)(nil)

// Hook represents the payload of a post-commit hook.
type Hook struct {
	Parent      int64                  `json:"parent"`
	Trigger     string                 `json:"trigger"`
	TriggeredBy int64                  `json:"triggered_by"`
	Link        string                 `json:"link"`
	Timestamp   int64                  `json:"timestamp"`
	Action      constant.TriggerAction `json:"action"`
	Ref         string                 `json:"ref"`
	Cron        string                 `json:"cron"`
	Sender      string                 `json:"sender"`
	Params      map[string]string      `json:"params"`
	// 重试相关字段
	RetrySourceID     *int64 `json:"retry_source_id,omitempty"`
	RetryCount        int    `json:"retry_count,omitempty"`
	RetrySourceNumber int64  `json:"retry_source_number,omitempty"`
}

// Triggerer is responsible for triggering a Execution from an
// incoming hook (could be manual or webhook). If an execution is skipped a nil value is
// returned.
type Triggerer interface {
	Trigger(ctx context.Context, pipelineID uint, hook *Hook) (*executionModel.Execution, error)
}

type triggerer struct{}

func New() Triggerer {
	return &triggerer{}
}

//nolint:gocognit,gocyclo,cyclop //TODO: Refactor @Vistaar
func (t *triggerer) Trigger(
	ctx context.Context,
	pipelineID uint,
	base *Hook,
) (*executionModel.Execution, error) {
	log := logger.GetLogger(ctx).With(
		"pipeline.id", pipelineID,
		"trigger.ref", base.Ref,
	)

	log.Info("trigger: received")
	defer func() {
		// taking the paranoid approach to recover from
		// a panic that should absolutely never happen.
		if r := recover(); r != nil {
			log.Errorf("runner: unexpected panic: %s", r)
			debug.PrintStack()
		}
	}()

	// TODO: implement trigger logic
	return nil, nil
}
