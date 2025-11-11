package repo

import (
	"context"
)

type Interface interface {
	Create(ctx context.Context, repo *Repository) error
	FindByRef(ctx context.Context, ref string) (*Repository, error)
	GetByID(ctx context.Context, id uint) (*Repository, error)
}
