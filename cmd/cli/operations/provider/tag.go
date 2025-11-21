package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/tag"
	tagSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/tag"
)

var TagWireSet = wire.NewSet(
	tag.NewController,
	wire.Bind(new(tagSvc.Interface), new(*tagSvc.Service)),
	tagSvc.NewService,
)
