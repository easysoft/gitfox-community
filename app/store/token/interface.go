package token

import (
	"context"
)

type Interface interface {
	GetByID(ctx context.Context, id int64) (*Token, error)
}
