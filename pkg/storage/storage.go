package storage

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver/factory"

	_ "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver/filesystem"
	_ "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver/inmemory"
)

func NewDriver(ctx context.Context, name string, parameters map[string]interface{}) (driver.StorageDriver, error) {
	return factory.Create(ctx, name, parameters)
}

type DriverGetter interface {
	Get(ctx context.Context, provider DriverType, config Config) (driver.StorageDriver, error)
}
