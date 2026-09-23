// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package storage

import (
	"context"

	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/storage/driver"
)

// WireSet 提供 storage 模块的依赖注入集合
var WireSet = wire.NewSet(
	ProvideStorageDriver,
)

// ProvideStorageDriver 根据配置创建存储驱动
func ProvideStorageDriver(cfg *config.Config) (driver.StorageDriver, func(), error) {
	storageCfg := cfg.Storage

	// 设置默认值
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

	// 构建驱动参数
	parameters := map[string]interface{}{
		"rootdirectory": rootDir,
		"maxthreads":    maxThreads,
	}

	ctx := context.Background()
	d, err := NewDriver(ctx, driverName, parameters)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		// 驱动清理逻辑（如果需要）
	}

	return d, cleanup, nil
}
