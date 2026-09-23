// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package user

import (
	"context"

	principal "github.com/easysoft/gitfox-community/app/model/user"
)

type Interface interface {
	FindByID(ctx context.Context, id int64) (*principal.User, error)
	FindByAccount(ctx context.Context, account string) (*principal.User, error)
	FindByEmail(ctx context.Context, email string) (*principal.User, error)
}

type CacheInterface interface {
	Find(ctx context.Context, account string) (*principal.User, error)
	FindMany(ctx context.Context, accounts []string) ([]*principal.User, error)
}
