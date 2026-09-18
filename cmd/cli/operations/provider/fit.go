package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/fit"
	fitSvc "github.com/easysoft/gitfox-community/app/service/fit"
)

var FitWireSet = wire.NewSet(
	fit.NewController,
	wire.Bind(new(fitSvc.Interface), new(*fitSvc.Service)),
	fitSvc.NewService,
)
