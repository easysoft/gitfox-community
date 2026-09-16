package storage

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/storage"
)

func LoadPipelineStorage(ctx context.Context, storageConfig storage.Config) (PipelineStorage, error) {
	var err error
	driver, err := storage.NewDriver(ctx, string(storageConfig.Local.Driver), storageConfig.Local.Parameters)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewCommonContentStore(ctx, driver, storage.WithPrefix(constant.CIPrefix))
	if err != nil {
		return nil, err
	}
	return &pipelineStorage{
		base: s,
	}, nil
}
