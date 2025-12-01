package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
	publickeySvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	tokenStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	userStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
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

	// Public Key Service (from publickey/wire.go)
	publickeySvc.WireSet,
)

// NewTokenStoreProvider creates a token store and returns it as Interface
func NewTokenStoreProvider(db *gorm.DB) tokenStore.Interface {
	return tokenStore.NewTokensStore(db)
}
