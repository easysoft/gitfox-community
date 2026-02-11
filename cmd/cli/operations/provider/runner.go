package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/runner"
	runnerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/runner"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
)

var RunnerWireSet = wire.NewSet(
	runner.NewController,
	// Runner service
	wire.Bind(new(runnerSvc.Interface), new(*runnerSvc.Service)),
	runnerSvc.NewService,
	// Runner store
	wire.Bind(new(runnerStore.Interface), new(*runnerStore.Store)),
	runnerStore.NewStore,
	// Runner token cache
	runnerStore.NewTokenCache,
)
