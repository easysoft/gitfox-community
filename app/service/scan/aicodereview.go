package scan

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
)

type AIReviewService struct {
	aireviewStore scanStore.IAIReview
}

var _ IAIReview = (*AIReviewService)(nil)

func NewService(aireviewStore scanStore.IAIReview) *AIReviewService {
	return &AIReviewService{
		aireviewStore: aireviewStore,
	}
}

func (s *AIReviewService) GetAIReviewRules(ctx context.Context) ([]*model.AIReviewRule, error) {
	rules, err := s.aireviewStore.GetRules(ctx)
	if err != nil {
		return rules, err
	}

	return rules, nil
}
