package scan

import (
	"context"
	"strings"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IAIReview = (*Store)(nil)

// Store aicodereview store implementation
type Store struct {
	db *gorm.DB
}

// NewAICodeReviewStore creates a new aicodereview store instance
func NewAICodeReviewStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetRules(ctx context.Context) ([]*model.AIReviewRule, error) {
	var rules []*AIReviewRule
	err := s.db.WithContext(ctx).Where("deleted = ?", false).Find(&rules).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	// 转换为 model 类型
	ruleModels := make([]*model.AIReviewRule, len(rules))
	for i, rule := range rules {
		extentions := strings.Split(strings.Trim(rule.Extention, ","), ",")
		for j, extention := range extentions {
			extentions[j] = "*." + extention
		}
		ruleModels[i] = &model.AIReviewRule{
			ID:        rule.ID,
			Language:  rule.Language,
			Extention: extentions,
			Rule:      rule.Rule,
		}
	}

	return ruleModels, nil
}
