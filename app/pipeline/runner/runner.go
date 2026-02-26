package runner

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/extend"
	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
)

type Runner interface {
	Run(ctx context.Context, stage *drone.Stage) error
}

func NewExecutionRunner(
	config *config.Config,
	client runnerclient.Client,
	stageEnvProvider *extend.StageDynamicEnvProvider,
	pipelineSvc pipeline.Interface,
) (Runner, error) {
	return newDockerRunner(config, client, stageEnvProvider, &pipelineSvc)
}
