// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
