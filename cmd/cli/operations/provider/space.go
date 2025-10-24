package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	spaceSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
)

var SpaceWireSet = wire.NewSet(
	space.NewSpaceController,
	wire.Bind(new(spaceSvc.SpaceService), new(*spaceSvc.SpaceSvc)),
	// 提供创建实现类实例的方法（关键修复）
	spaceSvc.NewSpaceSvc,
	wire.Bind(new(spaceStore.SpaceInterface), new(*spaceStore.MySQLSpace)),
	spaceStore.NewMySQLSpace,
)
