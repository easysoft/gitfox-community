package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"github.com/google/wire"
)

var ExecutionWireSet = wire.NewSet(
	execution.NewController,
)
