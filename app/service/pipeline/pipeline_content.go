package pipeline

import (
	"context"
	"encoding/json"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
)

// GetPipelineByID 根据流水线ID获取流水线（deleted=0）
func (s *Service) GetPipelineByID(ctx context.Context, id int64) (*pipelineModel.Pipeline, error) {
	return s.pipelineStore.GetPipelineByID(ctx, id)
}

func (s *Service) GetPipelineContent(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineFile, error) {
	pipelineFile, err := s.pipelineStore.GetPipelineContentByPipelineID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	return &pipelineModel.PipelineFile{
		Data: pipelineFile.Data,
	}, nil
}

// GetPipelineVariables 根据流水线ID获取流水线变量
// 如果 runtime=true，取 key 和 value；如果 runtime=false，取 key 和 defaultValue
func (s *Service) GetPipelineVariables(ctx context.Context, pipelineID int64) (map[string]string, error) {
	pipelineContent, err := s.pipelineStore.GetPipelineContentByPipelineID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	if pipelineContent.Variables == "" {
		return result, nil
	}

	var variables []pipelineModel.PipelineVariable
	if err := json.Unmarshal([]byte(pipelineContent.Variables), &variables); err != nil {
		return nil, err
	}

	for _, v := range variables {
		if v.Runtime {
			result[v.Key] = v.DefaultValue
		} else {
			result[v.Key] = v.Value
		}
	}

	return result, nil
}
