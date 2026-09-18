package stage

import (
	"context"

	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
)

// Create creates a new stage
func (s *Service) Create(ctx context.Context, stage *stageModel.Stage) error {
	return s.stageStore.Create(ctx, stage)
}

// Update updates a stage
func (s *Service) Update(ctx context.Context, stage *stageModel.Stage) error {
	return s.stageStore.Update(ctx, stage)
}

// GetByID gets stage by ID
func (s *Service) GetByID(ctx context.Context, id uint) (*stageModel.Stage, error) {
	return s.stageStore.GetByID(ctx, id)
}

func (s *Service) GetByIDWithSteps(ctx context.Context, id uint) (*stageModel.StageWithSteps, error) {
	return s.stageStore.GetByIDWithSteps(ctx, id)
}

// ListIncomplete returns a stage list where the stage is incomplete (pending or running)
func (s *Service) ListIncomplete(ctx context.Context) ([]*stageModel.Stage, error) {
	return s.stageStore.ListIncomplete(ctx)
}

// List returns a list of stages corresponding to an execution ID
func (s *Service) List(ctx context.Context, executionID uint) ([]*stageModel.Stage, error) {
	return s.stageStore.List(ctx, executionID)
}

// ListWithSteps returns a stage list with steps for an execution
func (s *Service) ListWithSteps(ctx context.Context, executionID uint) ([]*stageModel.StageWithSteps, error) {
	return s.stageStore.ListWithSteps(ctx, executionID)
}
