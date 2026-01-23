package pipeline

import "context"

type PluginService interface {
	Migrate(ctx context.Context, url string) error
}
