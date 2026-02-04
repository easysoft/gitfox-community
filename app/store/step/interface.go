package step

import (
	"context"

	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store step store implementation
type Store struct {
	db *gorm.DB
}

// NewStepStore creates a new step store instance
func NewStepStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
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
