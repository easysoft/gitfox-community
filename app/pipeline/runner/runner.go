// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package runner

import (
	"context"

	"github.com/easysoft/gitfox-community/app/pipeline/resolver"
	"github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/runner/extend"
	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
)

type Runner interface {
	Run(ctx context.Context, stage *drone.Stage) error
}

func NewExecutionRunner(
	config *config.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
	pipelineSvc pipeline.Interface,
) (Runner, error) {
	return newDockerRunner(config, client, resolver, stageEnvProvider, &pipelineSvc)
}
