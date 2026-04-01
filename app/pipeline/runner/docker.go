package runner

import (
	goruntime "runtime"

	runnerclient "github.com/drone/runner-go/client"

	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/pipeline/reporter/history"
	"github.com/drone/runner-go/pipeline/reporter/remote"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/pipeline/uploader"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/resolver"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine/compiler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine/linter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine/resource"

	compiler2 "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/compiler"
	engine2 "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/engine"

	runtime2 "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/runtime"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/extend"
)

// Privileged provides a list of plugins that execute
// with privileged capabilities in order to run Docker
// in Docker.
var Privileged = []string{
	"plugins/docker",
	"plugins/acr",
	"plugins/ecr",
	"plugins/gcr",
	"plugins/heroku",
}

// // dockerOpts returns back the options to be overridden from docker options set
// // in the environment. If values are specified in gitness, they get preference.
// func dockerOpts(config *types.Config) []dockerclient.Opt {
// 	var overrides []dockerclient.Opt
// 	if config.Docker.Host != "" {
// 		overrides = append(overrides, dockerclient.WithHost(config.Docker.Host))
// 	}
// 	if config.Docker.APIVersion != "" {
// 		overrides = append(overrides, dockerclient.WithVersion(config.Docker.APIVersion))
// 	}
// 	return overrides
// }

func newDockerRunner(
	config *config.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
	pipelineSvc *pipeline.Interface,
) (*runtime2.Runner, error) {
	// For linux/windows, containers need to have extra hosts set in order to interact with
	// gitness. For docker desktop for mac, this is built in and not needed.
	extraHosts := []string{}
	if goruntime.GOOS != "darwin" {
		extraHosts = []string{"host.docker.internal:host-gateway"}
	}

	compiler := &compiler.Compiler{
		Environ:    provider.Static(map[string]string{}),
		Registry:   registry.Static([]*drone.Registry{}),
		Secret:     secret.Encrypted(),
		ExtraHosts: extraHosts,
		Privileged: Privileged,
		Networks:   config.Ci.ContainerNetworks,
	}

	remote := remote.New(client)
	upload := uploader.New(client)
	tracer := history.New(remote)
	engine, err := engine.NewEnv(engine.Opts{APIVersion: config.Docker.ApiVersion})
	if err != nil {
		return nil, err
	}

	exec := runtime.NewExecer(tracer, remote, upload,
		engine, int64(config.Ci.ParallelWorkers))

	legacyRunner := &runtime.Runner{
		Machine:  config.InstanceID,
		Client:   client,
		Reporter: tracer,
		Lookup:   resource.Lookup,
		Lint:     linter.New().Lint,
		Compiler: compiler,
		Exec:     exec.Exec,
	}

	engine2, err := engine2.NewEnv(engine2.Opts{APIVersion: config.Docker.ApiVersion})
	if err != nil {
		return nil, err
	}

	exec2 := runtime2.NewExecer(tracer, remote, upload, engine2, int64(config.Ci.ParallelWorkers))

	compiler2 := &compiler2.CompilerImpl{
		Environ:     provider.Combine(provider.Static(map[string]string{}), stageEnvProvider),
		PipelineSvc: pipelineSvc,
		Secret:      secret.Encrypted(),
		ExtraHosts:  extraHosts,
		Privileged:  Privileged,
		Networks:    config.Ci.ContainerNetworks,
		Mount:       config.Ci.ContainerVolume,
	}

	runner := &runtime2.Runner{
		Machine:      config.InstanceID,
		Resolver:     resolver.GetLookupFn(),
		Client:       client,
		Reporter:     tracer,
		Compiler:     compiler2,
		Exec:         exec2.Exec,
		LegacyRunner: legacyRunner,
	}

	return runner, nil
}
