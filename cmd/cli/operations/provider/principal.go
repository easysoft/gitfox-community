package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	principalStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"github.com/google/wire"
)

var PrincipalWireSet = wire.NewSet(
	user.NewController,
	wire.Bind(new(userSvc.Interface), new(*userSvc.Service)),
	userSvc.NewService,
	wire.Bind(new(principalStore.Interface), new(*principalStore.Store)),
	principalStore.NewPrincipalStore,
)
