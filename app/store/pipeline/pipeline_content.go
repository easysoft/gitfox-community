package pipeline

import (
	"context"
	"encoding/base64"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// GetPipelineContentByPipelineID 根据流水线ID获取流水线内容
func (s *Store) GetPipelineContentByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineContent, error) {
	var content PipelineContent
	if err := s.db.WithContext(ctx).Where("pipelineID = ?", pipelineID).First(&content).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineContentToModel(content), nil
}

// mapPipelineContentToModel 将数据库模型映射为流水线内容模型
func mapPipelineContentToModel(content PipelineContent) *pipelineModel.PipelineContent {
	// 数据库中 data 字段存储的是 base64 编码后的字符串
	// 需要解码成原始字节数据供后续使用
	decodedData, err := base64.StdEncoding.DecodeString(string(content.Data))
	if err != nil {
		// 如果解码失败，可能数据库中存储的就是原始数据，直接使用
		decodedData = content.Data
	}

	return &pipelineModel.PipelineContent{
		ID:          content.ID,
		PipelineID:  content.PipelineID,
		Version:     content.Version,
		Data:        decodedData,              // 解码后的原始字节数据
		DataStr:     string(content.Data),     // 数据库中的 base64 字符串
		Variables:   content.Variables,
		CreatedBy:   content.CreatedBy,
		CreatedDate: content.CreatedDate,
		EditedBy:    content.EditedBy,
		EditedDate:  content.EditedDate,
	}
}
