package provider

import (
	runnerclient "github.com/drone/runner-go/client"
	"github.com/drone/runner-go/poller"
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/resolver"
	dockerRunner "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/storage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	runnerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/runner"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/extend"
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

// DockerRunnerWireSet provides a wire set for this package.
var DockerRunnerWireSet = wire.NewSet(
	ProvideDynamicStageProvider,
	ProvideResolverManager,
	ProvideExecutionRunner,
	ProvideExecutionPoller,
)

// ProvideDynamicStageProvider provides an stage dynamic env provider.
func ProvideDynamicStageProvider(
	store storage.PipelineStorage,
	stageStore stageStore.Interface,
) *extend.StageDynamicEnvProvider {
	return extend.NewStageDynamicEnvProvider(store, stageStore)
}

func ProvideResolverManager(
	pipelineSvc *pipeline.Service,
) *resolver.Manager {
	return resolver.NewManager(pipelineSvc)
}

// ProvideExecutionRunner provides an execution runner.
func ProvideExecutionRunner(
	config *config.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
	pipelineSvc *pipeline.Service,
) (dockerRunner.Runner, error) {
	return dockerRunner.NewExecutionRunner(config, client, resolver, stageEnvProvider, pipelineSvc)
}

// ProvideExecutionPoller provides a poller which can poll the manager
// for new builds and execute them.
func ProvideExecutionPoller(
	runner dockerRunner.Runner,
	client runnerclient.Client,
	rStore runnerStore.Interface,
	cfg *config.Config,
) *poller.Poller {
	return dockerRunner.NewExecutionPoller(runner, client, rStore, cfg)
}
