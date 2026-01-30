package pipeline

import (
	"context"
	"strings"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"
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

// GetTriggers 获取触发列表
func (s *Store) GetTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) ([]*pipelineModel.Trigger, error) {
	var triggers []Trigger
	query := s.db.WithContext(ctx).Where("deleted = 0")
	if filter.RepoID > 0 {
		query = query.Where("repoID = ?", filter.RepoID)
	}

	if filter.PipelineID > 0 {
		query = query.Where("pipelineID = ?", filter.PipelineID)
	}

	if err := query.Find(&triggers).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*pipelineModel.Trigger, 0, len(triggers))
	for _, t := range triggers {
		if t.Trigger != "" && len(filter.Trigger) > 0 {
			if len(slice.Intersection(filter.Trigger, strings.Split(t.Trigger, ","))) == 0 {
				continue
			}
		}
		result = append(result, mapTriggerToModel(&t))
	}

	return result, nil
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
		RepoID:      t.RepoID,
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
		RepoID:      t.RepoID,
		PipelineID:  t.PipelineID,
		Trigger:     strings.Join(t.Trigger, ","),
		Cron:        t.Cron,
		CreatedBy:   t.CreatedBy,
		CreatedDate: t.CreatedDate,
		EditedBy:    t.EditedBy,
		EditedDate:  t.EditedDate,
	}
}
