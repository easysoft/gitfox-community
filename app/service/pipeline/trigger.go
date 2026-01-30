package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// CreateTrigger 创建触发
func (s *Service) CreateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	return s.pipelineStore.CreateTrigger(ctx, trigger)
}

// UpdateTrigger 更新触发
func (s *Service) UpdateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	return s.pipelineStore.UpdateTrigger(ctx, trigger)
}

// GetTriggerByID 根据触发ID获取触发
func (s *Service) GetTriggerByID(ctx context.Context, id int64) (*pipelineModel.Trigger, error) {
	return s.pipelineStore.GetTriggerByID(ctx, id)
}

// GetTriggerByPipelineID 根据流水线ID获取触发
func (s *Service) GetTriggerByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.Trigger, error) {
	return s.pipelineStore.GetTriggerByPipelineID(ctx, pipelineID)
}

// ListTriggers 查询触发列表
func (s *Service) ListTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) (*database.PageResult[*pipelineModel.Trigger], error) {
	return s.pipelineStore.ListTrigger(ctx, filter)
}

// DeleteTrigger 删除触发
func (s *Service) DeleteTrigger(ctx context.Context, id int64) error {
	return s.pipelineStore.DeleteTrigger(ctx, id)
}
