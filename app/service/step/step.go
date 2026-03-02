package step

import (
	"context"

	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
)

// Create creates a new step
func (s *Service) Create(ctx context.Context, step *stepModel.Step) error {
	return s.stepStore.Create(ctx, step)
}

// Update updates a step
func (s *Service) Update(ctx context.Context, step *stepModel.Step) error {
	return s.stepStore.Update(ctx, step)
}

// GetByID gets step by ID
func (s *Service) GetByID(ctx context.Context, id uint) (*stepModel.Step, error) {
	return s.stepStore.GetByID(ctx, id)
}
