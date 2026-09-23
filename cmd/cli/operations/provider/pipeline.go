// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package provider

import (
	"context"

	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/pipeline"
	"github.com/easysoft/gitfox-community/app/constant"
	ciStorage "github.com/easysoft/gitfox-community/app/pipeline/storage"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	pipelineStore "github.com/easysoft/gitfox-community/app/store/pipeline"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/livelog"
	pkgS3 "github.com/easysoft/gitfox-community/pkg/s3"
	"github.com/easysoft/gitfox-community/pkg/storage"
	"github.com/easysoft/gitfox-community/pkg/storage/driver"
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

func ProvidePipelineCIStorage(cfg *config.Config) (ciStorage.PipelineStorage, error) {
	ctx := context.Background()

	storageCfg := cfg.Storage
	driverName := storageCfg.StorageType
	if driverName == "" {
		driverName = "filesystem"
	}

	rootDir := storageCfg.Filesystem.RootDirectory
	if rootDir == "" {
		rootDir = "/data/gitfox/storage"
	}

	maxThreads := storageCfg.Filesystem.MaxThreads
	if maxThreads == 0 {
		maxThreads = 100
	}

	storageConfig := storage.Config{
		Local: storage.DriverConfig{
			Driver:     storage.DriverType(driverName),
			Parameters: map[string]interface{}{
				"rootdirectory": rootDir,
				"maxthreads":    maxThreads,
			},
		},
	}

	return ciStorage.LoadPipelineStorage(ctx, storageConfig)
}
