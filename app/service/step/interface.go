package step

import (
	"context"

	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/step"
)

var _ Interface = (*Service)(nil)

type Service struct {
	stepStore step.Interface
}

func NewStepService(stepStore step.Interface) *Service {
	return &Service{
		stepStore: stepStore,
	}
}

type Interface interface {
	// Create creates a new step
	Create(ctx context.Context, step *stepModel.Step) error
	// Update updates a step
	Update(ctx context.Context, step *stepModel.Step) error
	// GetByID gets step by ID
	GetByID(ctx context.Context, id uint) (*stepModel.Step, error)
}
