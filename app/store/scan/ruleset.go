package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IRuleSet = (*RuleSetStore)(nil)

type RuleSetStore struct {
	db *gorm.DB
}

func NewRuleSetStore(db *gorm.DB) *RuleSetStore {
	return &RuleSetStore{
		db: db,
	}
}

func (s *RuleSetStore) Create(ctx context.Context, ruleSet *RuleSet) (*scanModel.RuleSet, error) {
	if err := s.db.WithContext(ctx).Create(&ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(*ruleSet), nil
}

func (s *RuleSetStore) GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error) {
	var ruleSet RuleSet
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(ruleSet), nil
}

func (s *RuleSetStore) Update(ctx context.Context, ruleSet *RuleSet, id uint) (*scanModel.RuleSet, error) {
	ruleSet.ID = int64(id)
	if err := s.db.WithContext(ctx).Updates(ruleSet).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 重新查询完整记录,因为 Updates 只更新指定非零值字段,不会回填其他字段
	var updated RuleSet
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(updated), nil
}

// mapToModel 将 RuleSet 转换为 scanModel.RuleSet
func mapToModel(data RuleSet) *scanModel.RuleSet {
	return &scanModel.RuleSet{
		ID:          data.ID,
		Name:        data.Name,
		Desc:        data.Desc,
		DescEn:      data.DescrEn,
		Lang:        data.Lang,
		Tag:         data.Tag,
		Status:      data.Status,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
		IsCustom:    data.IsCustom,
		Deleted:     data.Deleted,
	}
}
