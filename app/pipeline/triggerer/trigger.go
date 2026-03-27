package triggerer

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"slices"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/manager"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/scheduler"

	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	executionStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/execution"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	v1yaml "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
	normalize "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec/parse/normalize"
	specresolver "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec/parse/resolver"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/inputs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/script"
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
	Commit      string                 `json:"commit"`
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

type triggerer struct {
	pipelineService pipelineSvc.Interface
	executionStore  executionStore.Interface
	stageStore      stageStore.Interface
	scheduler       scheduler.Scheduler
	tx              dbtx.TxRunner
}

func New(pipelineService pipelineSvc.Interface, executionStore executionStore.Interface, stageStore stageStore.Interface, scheduler scheduler.Scheduler, tx dbtx.TxRunner) Triggerer {
	return &triggerer{
		pipelineService: pipelineService,
		executionStore:  executionStore,
		stageStore:      stageStore,
		scheduler:       scheduler,
		tx:              tx,
	}
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

	log.Info("[trigger]: received")
	defer func() {
		// taking the paranoid approach to recover from
		// a panic that should absolutely never happen.
		if r := recover(); r != nil {
			log.Errorf("[trigger]: unexpected panic: %s", r)
			debug.PrintStack()
		}
	}()

	// 从数据库查询 pipeline yaml 内容
	file, err := t.pipelineService.GetPipelineContent(ctx, int64(pipelineID))
	if err != nil {
		log.Errorf("[trigger]: could not find pipeline content: %v", err)
		return nil, err
	}

	tigger := string(base.Action)
	if tigger == "" {
		tigger = "manual"
	}
	execution := &executionModel.Execution{
		Commit:     base.Commit,
		PipelineID: pipelineID,
		Trigger:    tigger, // 使用 action 取代原先的 event
		CreatedBy:  base.Trigger,
		Parent:     uint(base.Parent),
		Status:     constant.CIStatusPending,
		Ref:        base.Ref,
		Params:     base.Params,
	}

	// For V1 YAML - for now we can just parse the stages and create them sequentially.
	stages, err := parseV1Stages(ctx, file.Data, execution, base, t.pipelineService)
	if err != nil {
		return nil, fmt.Errorf("could not parse v1 YAML into stages: %w", err)
	}

	if err := t.createExecutionWithStagesWithTx(ctx, execution, stages); err != nil {
		return nil, fmt.Errorf("could not create execution: %w", err)
	}

	for _, stage := range stages {
		if stage.Status != constant.CIStatusPending {
			continue
		}
		log.Infof("[trigger] scheduling stage code=%s name=%s", stage.Code, stage.Name)
		err = t.scheduler.Schedule()
		if err != nil {
			log.Errorf("[trigger]: cannot enqueue execution: %s", err)
			return nil, err
		}
	}

	return execution, nil
}

// parseV1Stages tries to parse the yaml into a list of stages and returns an error
// if we are unable to do so or the yaml contains something unexpected.
// Currently, all the stages will be executed one after the other on completion.
// Once we have depends on in v1, this will be changed to use the DAG.
//
//nolint:gocognit // refactor if needed.
func parseV1Stages(
	ctx context.Context,
	data []byte,
	execution *executionModel.Execution,
	base *Hook,
	pipelineService pipelineSvc.Interface,
) ([]*stageModel.Stage, error) {
	stages := []*stageModel.Stage{}
	// For V1 YAML, just go through the YAML and create stages serially for now
	config, err := v1yaml.ParseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse v1 yaml: %w", err)
	}

	// Normalize the config to make sure stage names and step names are unique
	err = normalize.Normalize(config)
	if err != nil {
		return nil, fmt.Errorf("could not normalize v1 yaml: %w", err)
	}

	if config.Kind != "pipeline" {
		return nil, fmt.Errorf("cannot support non-pipeline kinds in v1 at the moment: %w", err)
	}

	inputParams := map[string]interface{}{}
	inputParams["build"] = inputs.Build(manager.ConvertToDroneBuild(execution))
	inputParams["parameters"] = base.Params

	var prevStage string

	// expand stage level plugins
	lookupFunc := func(name, kind, typ, version string) (*v1yaml.Config, error) {
		// Call the pipeline service to resolve the plugin
		yamlConfig, err := pipelineService.Resolve(ctx, kind, typ, name)
		if err != nil {
			return nil, err
		}
		// Convert pipelineModel.Yaml to v1yaml.Config via JSON
		return ConvertToV1Config(yamlConfig)
	}

	if err := specresolver.Resolve(config, lookupFunc); err != nil {
		return nil, fmt.Errorf("could not resolve yaml plugins/templates: %w", err)
	}

	switch v := config.Spec.(type) {
	case *v1yaml.Pipeline:
		// Expand expressions in strings and matrices
		script.ExpandConfig(config, inputParams)

		runPolicyList := constant.RunPolicyList
		for idx, stage := range v.Stages {
			// Only parse CI stages for now
			switch stage.Spec.(type) {
			case *v1yaml.StageCI:
				runPolicy := string(constant.RunPolicyOnSuccess)
				if stage.When != nil {
					if when := stage.When.Eval; when != "" {
						trimmedWhen := strings.TrimSpace(when)
						if slices.Contains(runPolicyList, trimmedWhen) {
							// Direct policy value, use it as-is
							runPolicy = trimmedWhen
						} else {
							// Evaluate as expression
							onSuccess, onFailure, err := script.EvalWhen(when, inputParams)
							if err != nil {
								return nil, fmt.Errorf("could not resolve when condition for stage: %w", err)
							}
							switch {
							case onSuccess && onFailure:
								runPolicy = string(constant.RunPolicyAlways)
							case onSuccess:
								runPolicy = string(constant.RunPolicyOnSuccess)
							case onFailure:
								runPolicy = string(constant.RunPolicyOnFailure)
							}
						}
					}
				}

				dependsOn := []string{}
				if prevStage != "" {
					dependsOn = append(dependsOn, prevStage)
				}
				status := constant.CIStatusWaitingOnDeps
				// If the stage has no dependencies, it can be picked up for execution.
				if len(dependsOn) == 0 {
					status = constant.CIStatusPending
				}
				temp := &stageModel.Stage{
					ExecutionID: execution.ID,
					Code:        stage.Id,
					Name:        stage.Name,
					SN:          idx + 1,
					Depends:     dependsOn,
					Status:      status,
					RunPolicy:   runPolicy,
				}

				// Inline fixStageDetail logic
				stageCI := stage.Spec.(*v1yaml.StageCI)
				if stageCI.Runtime != nil {
					if stageCI.Runtime.Type == "agent" {
						temp.Type = "agent"
						if stageCI.Runtime.Spec != nil {
							runtimeSpec := stageCI.Runtime.Spec.(*v1yaml.RuntimeAgent)
							if len(runtimeSpec.Selector) > 0 {
								temp.Labels = runtimeSpec.Selector
							}
						}
					}
				}
				if stageCI.Platform != nil {
					if stageCI.Platform.Os != "" {
						temp.OS = stageCI.Platform.Os
					}
					if stageCI.Platform.Arch != "" {
						temp.Arch = stageCI.Platform.Arch
					}
				}

				prevStage = temp.Code
				stages = append(stages, temp)
			default:
				return nil, fmt.Errorf("only CI and template stages are supported in v1 at the moment")
			}
		}
	default:
		return nil, fmt.Errorf("unknown yaml: %w", err)
	}
	return stages, nil
}

// convertToV1Config converts pipelineModel.Yaml to v1yaml.Config via JSON serialization.
func ConvertToV1Config(src *pipelineModel.Yaml) (*v1yaml.Config, error) {
	if src == nil {
		return nil, fmt.Errorf("source yaml config is nil")
	}
	// Marshal pipelineModel.Yaml to JSON
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yaml config: %w", err)
	}
	// Unmarshal JSON to v1yaml.Config
	dst := &v1yaml.Config{}
	if err := json.Unmarshal(data, dst); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to v1yaml.Config: %w", err)
	}
	return dst, nil
}

// createExecutionWithStages writes an execution along with its stages in a single transaction.
func (t *triggerer) createExecutionWithStages(
	ctx context.Context,
	execution *executionModel.Execution,
	stages []*stageModel.Stage,
) error {
	err := t.executionStore.Create(ctx, execution)
	if err != nil {
		return err
	}

	for _, stage := range stages {
		stage.ExecutionID = execution.ID
		err := t.stageStore.Create(ctx, stage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *triggerer) createExecutionWithStagesWithTx(
	ctx context.Context,
	execution *executionModel.Execution,
	stages []*stageModel.Stage,
) error {
	return t.tx.WithTx(ctx, func(ctx context.Context) error {
		return t.createExecutionWithStages(ctx, execution, stages)
	})
}
