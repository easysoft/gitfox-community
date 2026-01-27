package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
)

func (s *Service) GetPipelineContent(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineFile, error) {
	pipelineFile, err := s.pipelineStore.GetPipelineContentByPipelineID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	return &pipelineModel.PipelineFile{
		Data: pipelineFile.Data,
	}, nil
}
