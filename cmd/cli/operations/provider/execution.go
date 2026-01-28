package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	"github.com/google/wire"
)

var ExecutionWireSet = wire.NewSet(
	execution.NewController,
	triggerer.New,
)
