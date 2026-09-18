package scan

import (
	"context"

	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
)

// NoopAIReview is the community-edition implementation of IAIReview.
//
// The AI code review feature is a paid capability. This no-op implementation
// keeps the public API surface and wiring intact while making it explicit that
// the feature is not available in this edition. The paid edition binds its own
// IAIReview implementation.
type NoopAIReview struct{}

var _ IAIReview = (*NoopAIReview)(nil)

// NewNoopAIReview creates the community-edition IAIReview implementation.
func NewNoopAIReview() *NoopAIReview {
	return &NoopAIReview{}
}

// GetAIReviewConfig always reports that the feature is unavailable.
func (s *NoopAIReview) GetAIReviewConfig(_ context.Context) (*scanModel.AIReviewConfig, error) {
	return nil, errors.BadRequest.New("AI code review is a paid feature and is not available in this edition")
}

// GetAIReviewRules always reports that the feature is unavailable.
func (s *NoopAIReview) GetAIReviewRules(_ context.Context) ([]*scanModel.AIReviewRule, error) {
	return nil, errors.BadRequest.New("AI code review is a paid feature and is not available in this edition")
}

// TriggerCodeReview is a no-op in the community edition. It deliberately
// returns (nil, nil) so that the pull request event flow is not disrupted and
// no spurious errors are logged.
func (s *NoopAIReview) TriggerCodeReview(_ context.Context, _ *CodeReviewInput) (*executionModel.Execution, error) {
	return nil, nil
}
