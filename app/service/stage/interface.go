package stage

import (
	"context"

	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
)

var _ Interface = (*Service)(nil)

type Service struct {
	stageStore stage.Interface
}

func NewStageService(stageStore stage.Interface) *Service {
	return &Service{
		stageStore: stageStore,
	}
}

type Interface interface {
	// Create creates a new stage
	Create(ctx context.Context, stage *stageModel.Stage) error
	// Update updates a stage
	Update(ctx context.Context, stage *stageModel.Stage) error
	// GetByID gets stage by ID
	GetByID(ctx context.Context, id uint) (*stageModel.Stage, error)
	// ListIncomplete returns a stage list where the stage is incomplete (pending or running)
	ListIncomplete(ctx context.Context) ([]*stageModel.Stage, error)
	// List returns a list of stages corresponding to an execution ID
	List(ctx context.Context, executionID uint) ([]*stageModel.Stage, error)
	// ListWithSteps returns a stage list with steps for an execution
	ListWithSteps(ctx context.Context, executionID uint) ([]*stageModel.StageWithSteps, error)
}
