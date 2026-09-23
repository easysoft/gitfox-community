// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/easysoft/gitfox-community/app/api/controller/user"
	userSvc "github.com/easysoft/gitfox-community/app/service/user"
	tokenStore "github.com/easysoft/gitfox-community/app/store/token"
	userStore "github.com/easysoft/gitfox-community/app/store/user"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var PrincipalWireSet = wire.NewSet(
	// User Controller
	user.NewUserController,

	// User Service
	wire.Bind(new(userSvc.Interface), new(*userSvc.Service)),
	userSvc.NewService,

	// Principal Store
	wire.Bind(new(userStore.Interface), new(*userStore.Store)),
	userStore.NewUserStore,

	// Token Store
	NewTokenStoreProvider,

	// Public Key Service
	PublicKeyWireSet,
)

// NewTokenStoreProvider creates a token store and returns it as Interface
func NewTokenStoreProvider(db *gorm.DB) tokenStore.Interface {
	return tokenStore.NewTokensStore(db)
}
