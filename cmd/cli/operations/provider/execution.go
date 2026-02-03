package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	executionStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/execution"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

var ExecutionWireSet = wire.NewSet(
	execution.NewController,
	triggerer.New,

	// Execution store
	wire.Bind(new(executionStore.Interface), new(*executionStore.Store)),
	executionStore.NewExecutionStore,

	// Stage store
	wire.Bind(new(stageStore.Interface), new(*stageStore.Store)),
	stageStore.NewStageStore,

	// Transaction runner
	wire.Bind(new(dbtx.TxRunner), new(*dbtx.TxRunnerImpl)),
	dbtx.NewTxRunner,
)
