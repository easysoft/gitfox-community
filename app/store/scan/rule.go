package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IRule = (*RuleStore)(nil)

type RuleStore struct {
	db *gorm.DB
}

func NewRuleStore(db *gorm.DB) *RuleStore {
	return &RuleStore{
		db: db,
	}
}

// ListByRuleSetID 根据规则集ID查询关联的规则列表
func (s *RuleStore) ListByRuleSetID(ctx context.Context, ruleIDs []int64) ([]Rule, error) {
	var rules []Rule

	if len(ruleIDs) == 0 {
		return rules, nil
	}

	err := s.db.WithContext(ctx).
		Where("id IN ? AND deleted = ?", ruleIDs, false).
		Find(&rules).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return rules, nil
}
