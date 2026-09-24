// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/storage"
	storagedriver "github.com/easysoft/gitfox-community/pkg/storage/driver"
)

// Config 保存 artifact 模块的配置。
// Config holds the artifact module configuration.
type Config struct {
	Storage ConfigStorage
}

type ConfigStorage struct {
	Prefix string
}

// LoadContentStorage 根据配置为 artifact 文件创建独立的 ContentStorage。
// LoadContentStorage creates a ContentStorage for artifact files.
func LoadContentStorage(ctx context.Context, driver storagedriver.StorageDriver, c Config) (storage.ContentStorage, error) {
	prefix := c.Storage.Prefix
	if prefix == "" {
		prefix = "artifact"
	}
	return storage.NewCommonContentStore(ctx, driver, storage.WithPrefix(prefix))
}
