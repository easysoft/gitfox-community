// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
