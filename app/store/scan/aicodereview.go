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
	extensionList := make([]string, 0)
	for i, rule := range rules {
		extensions := strings.Split(strings.Trim(rule.Extension, ","), ",")
		for _, extension := range extensions {
			if extension != "" {
				extensionList = append(extensions, "*."+extension)
			}
		}
		ruleModels[i] = &model.AIReviewRule{
			ID:        rule.ID,
			Language:  rule.Language,
			Extension: extensionList,
			Rule:      rule.Rule,
		}
	}

	return ruleModels, nil
}
