package extend

import (
	"bytes"
	"context"
	"errors"

	"github.com/easysoft/gitfox-community/app/pipeline/storage"
	stageStore "github.com/easysoft/gitfox-community/app/store/stage"
	schema "github.com/easysoft/gitfox-community/pkg/pipeline/spec"

	"github.com/drone/runner-go/environ/provider"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

type StageDynamicEnvProvider struct {
	store      storage.PipelineStorage
	stageStore stageStore.Interface
}

func NewStageDynamicEnvProvider(store storage.PipelineStorage, stageStore stageStore.Interface) *StageDynamicEnvProvider {
	return &StageDynamicEnvProvider{
		store:      store,
		stageStore: stageStore,
	}
}

func (p *StageDynamicEnvProvider) List(ctx context.Context, in *provider.Request) ([]*provider.Variable, error) {
	logger.GetLogger(ctx).Debug("start to list stage dynamicEnvs")
	if p.store == nil {
		return nil, nil
	}

	stageCtx := FromContext(ctx)
	if stageCtx == nil {
		return nil, nil
	}

	stageSpec, ok := stageCtx.stage.Spec.(*schema.StageCI)
	if !ok {
		return nil, errors.New("ci stage resource not found")
	}

	var dynamicEnvs = make([]*provider.Variable, 0)
	if len(stageSpec.Vars) == 0 {
		return dynamicEnvs, nil
	}
	for _, varObj := range stageSpec.Vars {
		if varObj.Type != "stage" {
			continue
		}

		srcStage, stageNum := loadSourceStage(varObj.Name, stageCtx.pipeline)
		if srcStage == nil {
			continue
		}

		dbStage, _ := p.stageStore.FindByNumber(ctx, in.Build.ID, stageNum)
		envBytes, e := p.store.GetStageFile(ctx, int64(dbStage.ExecutionID), int64(dbStage.ID), ".custom_env")
		if e != nil {
			return nil, e
		}

		for _, line := range bytes.Split(envBytes, []byte("\n")) {
			i := bytes.Index(line, []byte("="))
			if i == -1 {
				continue
			}
			key := string(line[:i])
			value := string(line[i+1:])
			dynamicEnvs = append(dynamicEnvs, &provider.Variable{
				Name: key, Data: value,
			})
			logger.GetLogger(ctx).Infof("add variable: %s", key)
		}
	}

	return dynamicEnvs, nil
}

func loadSourceStage(key string, in *schema.Pipeline) (*schema.Stage, int) {
	var s *schema.Stage
	var number int
	for idx, stage := range in.Stages {
		if stage.Name == key {
			s = stage
			number = idx + 1
			break
		}
	}
	return s, number
}
