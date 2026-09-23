// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package settings

import (
	"context"

	"github.com/easysoft/gitfox-community/app/model/settings"
)

type Interface interface {
	GetSpaceSettings(ctx context.Context, spaceID uint) ([]*settings.Setting, error)
	GetRepoSettings(ctx context.Context, repoID uint) ([]*settings.Setting, error)
	GetSystemSettings(ctx context.Context) ([]*settings.SystemSetting, error)

	GetSpaceSettingByKey(ctx context.Context, spaceID uint, key string) (*settings.Setting, error)
	GetRepoSettingByKey(ctx context.Context, repoID uint, key string) (*settings.Setting, error)
	GetSystemSettingByKey(ctx context.Context, key string) (*settings.SystemSetting, error)
}
