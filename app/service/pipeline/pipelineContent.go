package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
)

func (s *Service) GetPipelineController(ctx context.Context, projectID int64) (*pipelineModel.PipelineFile, error) {
	pipelineFile, err := s.pipelineStore.GetPipelineContentByPipelineID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &pipelineModel.PipelineFile{
		Data: pipelineFile.Data,
	}, nil
}
