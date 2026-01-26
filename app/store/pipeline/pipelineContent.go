package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// GetPipelineContentByPipelineID 根据流水线ID获取流水线内容
func (s *Store) GetPipelineContentByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineContent, error) {
	var content PipelineContent
	if err := s.db.WithContext(ctx).Where("pipelineID = ? AND deleted = 0", pipelineID).First(&content).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineContentToModel(content), nil
}

// mapPipelineContentToModel 将数据库模型映射为流水线内容模型
func mapPipelineContentToModel(content PipelineContent) *pipelineModel.PipelineContent {
	return &pipelineModel.PipelineContent{
		ID:          content.ID,
		PipelineID:  content.PipelineID,
		Version:     content.Version,
		Data:        content.Data,
		DataStr:     string(content.Data),
		Variables:   content.Variables,
		CreatedBy:   content.CreatedBy,
		CreatedDate: content.CreatedDate,
		EditedBy:    content.EditedBy,
		EditedDate:  content.EditedDate,
	}
}
