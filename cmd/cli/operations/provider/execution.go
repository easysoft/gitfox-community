package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/canceler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/manager"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/scheduler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	executionSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	stageSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	stepSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"
	executionStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/execution"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	stepStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

var ExecutionWireSet = wire.NewSet(
	execution.NewController,
	triggerer.New,
	scheduler.New,
	canceler.New,
	wire.Bind(new(manager.ExecutionManager), new(*manager.Manager)),
	manager.New,
	manager.NewEmbeddedClient,

	// Execution service
	wire.Bind(new(executionSvc.Interface), new(*executionSvc.Service)),
	executionSvc.NewExecutionService,
	// Execution store
	wire.Bind(new(executionStore.Interface), new(*executionStore.Store)),
	executionStore.NewExecutionStore,

	// Stage service
	wire.Bind(new(stageSvc.Interface), new(*stageSvc.Service)),
	stageSvc.NewStageService,
	// Stage store
	wire.Bind(new(stageStore.Interface), new(*stageStore.Store)),
	stageStore.NewStageStore,

	// Step service
	wire.Bind(new(stepSvc.Interface), new(*stepSvc.Service)),
	stepSvc.NewStepService,
	// Step store
	wire.Bind(new(stepStore.Interface), new(*stepStore.Store)),
	stepStore.NewStepStore,

	// Transaction runner
	wire.Bind(new(dbtx.TxRunner), new(*dbtx.TxRunnerImpl)),
	dbtx.NewTxRunner,
)
