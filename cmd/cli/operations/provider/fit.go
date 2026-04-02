package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/fit"
	fitSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/fit"
)

var FitWireSet = wire.NewSet(
	fit.NewController,
	wire.Bind(new(fitSvc.Interface), new(*fitSvc.Service)),
	fitSvc.NewService,
)
