// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"context"
	"encoding/json"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
)

// GetPipelineByID 根据流水线ID获取流水线（deleted=0）
func (s *Service) GetPipelineByID(ctx context.Context, id int64) (*pipelineModel.Pipeline, error) {
	return s.pipelineStore.GetPipelineByID(ctx, id)
}

// UpdatePipeline 更新流水线
func (s *Service) UpdatePipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) error {
	return s.pipelineStore.UpdatePipeline(ctx, pipeline)
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
