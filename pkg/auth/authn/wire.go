// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package authn

import (
	"github.com/easysoft/gitfox-community/app/store/token"
	principal "github.com/easysoft/gitfox-community/app/store/user"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/google/wire"
)

var AuthnWireSet = wire.NewSet(
	ProvideAuthenticators,
)

func ProvideAuthenticators(
	cfg *config.Config,
	principalStore principal.Interface,
	tokenStore token.Interface,
) Authenticator {
	basic := NewBasicAuthenticator(principalStore, cfg.PipelineAuth)
	_ = Register(AuthKindBasic, basic)

	easy := NewEasyAuthenticator(cfg, principalStore, tokenStore)
	_ = Register(AuthKindEasy, easy)

	return easy
}
