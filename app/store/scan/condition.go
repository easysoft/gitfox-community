package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IScanCondition = (*ScanConditionStore)(nil)
var _ IScanPlanCondition = (*ScanConditionStore)(nil)

type ScanConditionStore struct {
	db *gorm.DB
}

func NewScanConditionStore(db *gorm.DB) *ScanConditionStore {
	return &ScanConditionStore{db: db}
}

func (s *ScanConditionStore) BatchCreate(ctx context.Context, conditions []*ScanPlanCondition) error {
	if len(conditions) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Create(&conditions).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanConditionStore) BatchDelete(ctx context.Context, repoID uint, planID uint, triggerID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanPlanCondition{}).
		Where("repoID = ? AND planID = ? AND triggerID = ? AND id IN ?", repoID, planID, triggerID, ids).
		Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanConditionStore) List(ctx context.Context, triggerID uint) ([]*scanModel.ScanCondition, error) {
	var conditions []*ScanPlanCondition
	if err := s.db.WithContext(ctx).Where("triggerID = ? AND deleted = ?", triggerID, false).Find(&conditions).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.ScanCondition, 0, len(conditions))
	for _, c := range conditions {
		result = append(result, mapScanConditionToModel(*c))
	}
	return result, nil
}

// ---- plan-scoped conditions (triggerID=0) ----

func (s *ScanConditionStore) Create(ctx context.Context, condition *ScanPlanCondition) (*scanModel.ScanCondition, error) {
	if err := s.db.WithContext(ctx).Create(condition).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanConditionToModel(*condition), nil
}

func (s *ScanConditionStore) GetByID(ctx context.Context, id uint) (*scanModel.ScanCondition, error) {
	var condition ScanPlanCondition
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&condition).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanConditionToModel(condition), nil
}

func (s *ScanConditionStore) Update(ctx context.Context, condition *ScanPlanCondition) error {
	if err := s.db.WithContext(ctx).Updates(condition).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanConditionStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&ScanPlanCondition{}).
		Where("id = ?", id).Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanConditionStore) ListByPlan(ctx context.Context, repoID, planID uint) ([]*scanModel.ScanCondition, error) {
	var conditions []*ScanPlanCondition
	if err := s.db.WithContext(ctx).
		Where("repoID = ? AND planID = ? AND triggerID = 0 AND deleted = ?", repoID, planID, false).
		Find(&conditions).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.ScanCondition, 0, len(conditions))
	for _, c := range conditions {
		result = append(result, mapScanConditionToModel(*c))
	}
	return result, nil
}

func (s *ScanConditionStore) BatchCreateByPlan(ctx context.Context, conditions []*ScanPlanCondition) error {
	// triggerID must already be set to 0 by caller.
	return s.BatchCreate(ctx, conditions)
}

func (s *ScanConditionStore) BatchDeleteByIDs(ctx context.Context, repoID, planID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Model(&ScanPlanCondition{}).
		Where("repoID = ? AND planID = ? AND triggerID = 0 AND id IN ?", repoID, planID, ids).
		Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func mapScanConditionToModel(data ScanPlanCondition) *scanModel.ScanCondition {
	return &scanModel.ScanCondition{
		ID:          int64(data.ID),
		PlanID:      int64(data.PlanID),
		TriggerID:   int64(data.TriggerID),
		RepoID:      int64(data.RepoID),
		Priority:    scanModel.ScanRulePriority(data.Priority),
		Type:        scanModel.ScanRuleType(data.Type),
		Unit:        scanModel.ScanRuleUnit(data.Unit),
		Threshold:   data.Threshold,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}
