package pipeline

import (
	"context"
	"strings"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// CreateTrigger 创建触发
func (s *Store) CreateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	storeTrigger := mapTriggerToStore(trigger)
	if err := s.db.WithContext(ctx).Create(&storeTrigger).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	trigger.ID = storeTrigger.ID
	trigger.CreatedDate = storeTrigger.CreatedDate
	trigger.EditedDate = storeTrigger.EditedDate
	return nil
}

// UpdateTrigger 更新触发，如果 cron 和 trigger 都为空则硬删除
func (s *Store) UpdateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	// 如果 cron 和 trigger 都为空，则硬删除
	if trigger.Cron == "" && len(trigger.Trigger) == 0 {
		return s.DeleteTrigger(ctx, trigger.ID)
	}

	storeTrigger := mapTriggerToStore(trigger)
	if err := s.db.WithContext(ctx).Model(&Trigger{}).Where("id = ? AND deleted = 0", trigger.ID).Updates(&storeTrigger).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetTriggerByID 根据触发ID获取触发
func (s *Store) GetTriggerByID(ctx context.Context, id int64) (*pipelineModel.Trigger, error) {
	var trigger Trigger
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapTriggerToModel(&trigger), nil
}

// GetTriggerByPipelineID 根据流水线ID获取触发
func (s *Store) GetTriggerByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.Trigger, error) {
	var trigger Trigger
	if err := s.db.WithContext(ctx).Where("pipelineID = ? AND deleted = 0", pipelineID).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapTriggerToModel(&trigger), nil
}

// ListTrigger 查询触发列表（分页）
func (s *Store) ListTrigger(ctx context.Context, filter *pipelineModel.TriggerFilter) (*database.PageResult[*pipelineModel.Trigger], error) {
	tx := s.db.WithContext(ctx).Model(&Trigger{}).Where("deleted = 0")

	// 流水线ID过滤
	if filter.PipelineID != 0 {
		tx = tx.Where("pipelineID = ?", filter.PipelineID)
	}

	// 触发类型过滤
	// - 如果传入 ["push", "tag"]
	// - 生成的 SQL 类似：WHERE (`trigger` LIKE '%push%' OR `trigger` LIKE '%tag%')
	// - 会匹配到 push、tag、push,tag、merge,push 等所有包含这些值的记录
	if len(filter.Trigger) > 0 {
		var orConditions []string
		var args []any
		for _, t := range filter.Trigger {
			orConditions = append(orConditions, "`trigger` LIKE ?")
			args = append(args, "%"+t+"%")
		}
		tx = tx.Where(strings.Join(orConditions, " OR "), args...)
	}

	// 执行分页查询
	var triggers []*Trigger
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &triggers)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 Model 层结构
	result := make([]*pipelineModel.Trigger, len(triggers))
	for i := range triggers {
		result[i] = mapTriggerToModel(triggers[i])
	}

	return &database.PageResult[*pipelineModel.Trigger]{
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
		Data:     result,
	}, nil
}

// DeleteTrigger 删除触发（硬删除）
func (s *Store) DeleteTrigger(ctx context.Context, id int64) error {
	if err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&Trigger{}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// mapTriggerToModel 将 Store 层结构转换为 Model 层结构
func mapTriggerToModel(t *Trigger) *pipelineModel.Trigger {
	var triggerList []string
	if t.Trigger != "" {
		triggerList = strings.Split(t.Trigger, ",")
	}
	return &pipelineModel.Trigger{
		ID:          t.ID,
		PipelineID:  t.PipelineID,
		Trigger:     triggerList,
		Cron:        t.Cron,
		CreatedBy:   t.CreatedBy,
		CreatedDate: t.CreatedDate,
		EditedBy:    t.EditedBy,
		EditedDate:  t.EditedDate,
	}
}

// mapTriggerToStore 将 Model 层结构转换为 Store 层结构
func mapTriggerToStore(t *pipelineModel.Trigger) *Trigger {
	return &Trigger{
		ID:          t.ID,
		PipelineID:  t.PipelineID,
		Trigger:     strings.Join(t.Trigger, ","),
		Cron:        t.Cron,
		CreatedBy:   t.CreatedBy,
		CreatedDate: t.CreatedDate,
		EditedBy:    t.EditedBy,
		EditedDate:  t.EditedDate,
	}
}
