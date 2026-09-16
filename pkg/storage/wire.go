// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

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
