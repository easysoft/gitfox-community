package notify

import (
	"context"
)

type Interface interface {
	Create(ctx context.Context, notify *Notify) error
	Update(ctx context.Context, id uint, status string) error
}
