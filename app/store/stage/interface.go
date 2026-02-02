package stage

import (
	"context"

	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store stage store implementation
type Store struct {
	db *gorm.DB
}

// NewStageStore creates a new stage store instance
func NewStageStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Interface interface {
	// Create creates a new stage
	Create(ctx context.Context, stage *stageModel.Stage) error
	// GetByID gets stage by ID
	GetByID(ctx context.Context, id uint) (*stageModel.Stage, error)
}
