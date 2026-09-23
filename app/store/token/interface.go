// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package token

import (
	"context"
	"time"

	"github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
)

type Interface interface {
	GetByID(ctx context.Context, id int64) (*Token, error)
	GetByIdentifier(ctx context.Context, account string, identifier string) (*Token, error)
	GetDevopsToken(ctx context.Context) (*Entry, error)
	SetDevopsToken(ctx context.Context) error
	UpdateToken(ctx context.Context, id int64, token string) error

	CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
	CreateAppPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error)
}
