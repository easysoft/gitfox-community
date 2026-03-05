package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IScanTrigger = (*ScanTriggerStore)(nil)

type ScanTriggerStore struct {
	db *gorm.DB
}

func NewScanTriggerStore(db *gorm.DB) *ScanTriggerStore {
	return &ScanTriggerStore{db: db}
}

func (s *ScanTriggerStore) Create(ctx context.Context, trigger *ScanPlanTrigger) (*scanModel.ScanTrigger, error) {
	if err := s.db.WithContext(ctx).Create(trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(*trigger), nil
}

func (s *ScanTriggerStore) GetByID(ctx context.Context, id uint) (*scanModel.ScanTrigger, error) {
	var trigger ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(trigger), nil
}

func (s *ScanTriggerStore) GetByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error) {
	var trigger ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("repoID = ? AND planID = ? AND id = ? AND deleted = ?", repoID, planID, id, false).First(&trigger).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTriggerToModel(trigger), nil
}

func (s *ScanTriggerStore) Update(ctx context.Context, trigger *ScanPlanTrigger) error {
	if err := s.db.WithContext(ctx).Updates(trigger).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanTriggerStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&ScanPlanTrigger{}).
		Where("id = ?", id).Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *ScanTriggerStore) List(ctx context.Context, planID uint) ([]*scanModel.ScanTrigger, error) {
	var triggers []*ScanPlanTrigger
	if err := s.db.WithContext(ctx).Where("planID = ? AND deleted = ?", planID, false).Find(&triggers).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*scanModel.ScanTrigger, 0, len(triggers))
	for _, t := range triggers {
		result = append(result, mapScanTriggerToModel(*t))
	}
	return result, nil
}

func mapScanTriggerToModel(data ScanPlanTrigger) *scanModel.ScanTrigger {
	return &scanModel.ScanTrigger{
		ID:          int64(data.ID),
		PlanID:      int64(data.PlanID),
		RepoID:      int64(data.RepoID),
		Name:        data.Name,
		Desc:        data.Desc,
		Type:        scanModel.ScanTriggerType(data.Type),
		Cron:        data.Cron,
		CronBranch:  data.CronBranch,
		Keywords:    data.Keywords,
		Actions:     data.Actions,
		Disabled:    data.Disabled,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		Deleted:     data.Deleted,
	}
}
