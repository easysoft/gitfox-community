package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
	publickeySvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	principalStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	tokenStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
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
	wire.Bind(new(principalStore.Interface), new(*principalStore.Store)),
	principalStore.NewPrincipalStore,

	// Token Store
	NewTokenStoreProvider,

	// Public Key Service (from publickey/wire.go)
	publickeySvc.WireSet,
)

// NewTokenStoreProvider creates a token store and returns it as Interface
func NewTokenStoreProvider(db *gorm.DB) tokenStore.Interface {
	return tokenStore.NewTokensStore(db)
}
