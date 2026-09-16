package execution

import (
	"context"

	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store  execution store implementation
type Store struct {
	db *gorm.DB
}

// NewExecutionStore creates a new execution store instance
func NewExecutionStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Interface interface {
	// Create creates a new execution
	Create(ctx context.Context, execution *executionModel.Execution) error
	// Update updates an execution
	Update(ctx context.Context, execution *executionModel.Execution) error
	// GetByID gets execution by ID
	GetByID(ctx context.Context, id uint) (*executionModel.Execution, error)
	// ListByPr gets executions by PR with pipeline name
	ListByPr(ctx context.Context, filter executionModel.ExecutionInfoReqFilter) ([]*executionModel.ExecutionByPr, error)
}
