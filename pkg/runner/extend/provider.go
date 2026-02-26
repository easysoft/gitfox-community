package extend

import (
	"bytes"
	"context"
	"errors"

	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	schema "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox/app/pipeline/storage"

	"github.com/drone/runner-go/environ/provider"
	"github.com/rs/zerolog/log"
)

type StageDynamicEnvProvider struct {
	store    storage.PipelineStorage
	stageSvc stage.Interface
}

func NewStageDynamicEnvProvider(store storage.PipelineStorage, stageSvc stage.Interface) *StageDynamicEnvProvider {
	return &StageDynamicEnvProvider{
		store:    store,
		stageSvc: stageSvc,
	}
}

func (p *StageDynamicEnvProvider) List(ctx context.Context, in *provider.Request) ([]*provider.Variable, error) {
	log.Ctx(ctx).Debug().Msgf("start to list stage dynamicEnvs")
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

		stageList, _ := p.stageSvc.List(ctx, uint(in.Build.ID))
		var stageInfo *stageModel.Stage
		for _, stage := range stageList {
			if stage.SN == stageNum {
				stageInfo = stage
			}
		}
		envBytes, e := p.store.GetStageFile(ctx, int64(stageInfo.ExecutionID), int64(stageInfo.ID), ".custom_env")
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
			log.Ctx(ctx).Info().Msgf("add variable: %s", key)
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
