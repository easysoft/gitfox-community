package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/tag"
	tagSvc "github.com/easysoft/gitfox-community/app/service/tag"
)

var TagWireSet = wire.NewSet(
	tag.NewController,
	wire.Bind(new(tagSvc.Interface), new(*tagSvc.Service)),
	tagSvc.NewService,
)
