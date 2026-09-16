package pipeline

import (
	"context"
	"encoding/base64"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/database"
)

// CreatePipelineContent 创建流水线内容
func (s *Store) CreatePipelineContent(ctx context.Context, content *pipelineModel.PipelineContent) (*pipelineModel.PipelineContent, error) {
	dbContent := mapModelToPipelineContent(content)
	if err := s.db.WithContext(ctx).Create(dbContent).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineContentToModel(*dbContent), nil
}

// UpdatePipelineContent 更新流水线内容
func (s *Store) UpdatePipelineContent(ctx context.Context, content *pipelineModel.PipelineContent) error {
	if err := s.db.WithContext(ctx).Where("pipelineID = ?", content.PipelineID).Updates(&PipelineContent{
		Data: content.Data, // 只更新 data 字段，其他字段保持不变
	}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetPipelineContentByPipelineID 根据流水线ID获取流水线内容
func (s *Store) GetPipelineContentByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineContent, error) {
	var content PipelineContent
	if err := s.db.WithContext(ctx).Where("pipelineID = ?", pipelineID).First(&content).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineContentToModel(content), nil
}

// mapModelToPipelineContent 将流水线内容模型映射为数据库模型
func mapModelToPipelineContent(content *pipelineModel.PipelineContent) *PipelineContent {
	return &PipelineContent{
		ID:         content.ID,
		PipelineID: content.PipelineID,
		Version:    content.Version,
		Data:       content.Data,
		Variables:  database.NoNull(content.Variables),
		CreatedBy:  content.CreatedBy,
	}
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
		Data:        decodedData,          // 解码后的原始字节数据
		DataStr:     string(content.Data), // 数据库中的 base64 字符串
		Variables:   content.Variables.String,
		CreatedBy:   content.CreatedBy,
		CreatedDate: content.CreatedDate,
		EditedBy:    content.EditedBy,
		EditedDate:  content.EditedDate,
	}
}
