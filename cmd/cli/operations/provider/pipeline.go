package provider

import (
	"context"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	ciStorage "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/storage"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	pkgS3 "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/s3"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage/driver"
)

var PipelineWireSet = wire.NewSet(
	pipeline.NewController,
	wire.Bind(new(pipelineSvc.Interface), new(*pipelineSvc.Service)),
	pipelineSvc.NewPipelineService,
	wire.Bind(new(pipelineStore.Interface), new(*pipelineStore.Store)),
	pipelineStore.NewPipelineStore,
	ProvidePipelineS3Store,
	ProvidePipelineContentStorage,
	ProvidePipelineCIStorage,
	livelog.ProvideLogStream,
)

// ProvideS3Store 根据配置决定是否创建 S3 存储实例，未开启则返回 nil
func ProvidePipelineS3Store(cfg *config.Config) (*pkgS3.Store, error) {
	if !cfg.Storage.S3.Enabled {
		return nil, nil
	}

	cfg.Storage.S3.Prefix = constant.PipelinePrefix
	return pkgS3.NewS3Store(&cfg.Storage.S3)
}

// ProvideContentStorage 创建内容存储服务
func ProvidePipelineContentStorage(cfg *config.Config, d driver.StorageDriver) (storage.ContentStorage, error) {
	ctx := context.Background()

	var opts []storage.CommonContentStoreOption
	opts = append(opts, storage.WithPrefix(constant.PipelinePrefix))
	return storage.NewCommonContentStore(ctx, d, opts...)
}

func ProvidePipelineCIStorage(ctx context.Context, storageConfig storage.Config) (ciStorage.PipelineStorage, error) {
	return ciStorage.LoadPipelineStorage(ctx, storageConfig)
}
