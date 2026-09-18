package storage

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/storage/driver"
	"github.com/easysoft/gitfox-community/pkg/storage/driver/factory"

	_ "github.com/easysoft/gitfox-community/pkg/storage/driver/filesystem"
	_ "github.com/easysoft/gitfox-community/pkg/storage/driver/inmemory"
)

func NewDriver(ctx context.Context, name string, parameters map[string]interface{}) (driver.StorageDriver, error) {
	return factory.Create(ctx, name, parameters)
}

type DriverGetter interface {
	Get(ctx context.Context, provider DriverType, config Config) (driver.StorageDriver, error)
}
