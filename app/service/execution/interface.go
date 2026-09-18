package execution

import (
	"context"

	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	"github.com/easysoft/gitfox-community/app/store/execution"
	"github.com/easysoft/gitfox-community/app/store/stage"
)

var _ Interface = (*Service)(nil)

type Service struct {
	executionStore execution.Interface
	stageStore     stage.Interface
}

func NewExecutionService(executionStore execution.Interface, stageStore stage.Interface) *Service {
	return &Service{
		executionStore: executionStore,
		stageStore:     stageStore,
	}
}

type Interface interface {
	// Create creates a new execution
	Create(ctx context.Context, execution *executionModel.Execution) error
	// Update updates an execution
	Update(ctx context.Context, execution *executionModel.Execution) error
	// GetByID gets execution by ID
	GetByID(ctx context.Context, id uint) (*executionModel.Execution, error)
	// GetWithStages gets execution with all stages and steps
	GetWithStages(ctx context.Context, id uint) (*executionModel.ExecutionWithStages, error)
	// ListByPr gets executions by PR with pipeline name
	ListByPr(ctx context.Context, filter executionModel.ExecutionInfoReqFilter) ([]*executionModel.ExecutionByPr, error)
}
