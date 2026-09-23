// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/artifact"
	"github.com/easysoft/gitfox-community/app/api/controller/branch"
	"github.com/easysoft/gitfox-community/app/api/controller/commit"
	"github.com/easysoft/gitfox-community/app/api/controller/cron"
	"github.com/easysoft/gitfox-community/app/api/controller/execution"
	"github.com/easysoft/gitfox-community/app/api/controller/fit"
	"github.com/easysoft/gitfox-community/app/api/controller/git"
	"github.com/easysoft/gitfox-community/app/api/controller/pipeline"
	"github.com/easysoft/gitfox-community/app/api/controller/repo"
	"github.com/easysoft/gitfox-community/app/api/controller/runner"
	"github.com/easysoft/gitfox-community/app/api/controller/scan"
	"github.com/easysoft/gitfox-community/app/api/controller/space"
	"github.com/easysoft/gitfox-community/app/api/controller/statistics"
	"github.com/easysoft/gitfox-community/app/api/controller/svn"
	"github.com/easysoft/gitfox-community/app/api/controller/tag"
	"github.com/easysoft/gitfox-community/app/api/controller/user"
	"github.com/easysoft/gitfox-community/app/api/controller/webhook"
	"github.com/easysoft/gitfox-community/app/api/router"
)

var ControllersWireSet = wire.NewSet(
	NewControllers,
)

// NewControllers 创建所有 HTTP 处理器的聚合实例
func NewControllers(
	spaceController *space.SpaceController,
	branchController *branch.Controller,
	pipelineController *pipeline.Controller,
	executionController *execution.Controller,
	runnerController *runner.Controller,
	svnController *svn.Controller,
	tagController *tag.Controller,
	commitController *commit.Controller,
	repoController *repo.Controller,
	userController *user.Controller,
	gitController *git.GitController,
	scanController *scan.ScanController,
	fitController *fit.Controller,
	webhookController *webhook.Controller,
	artifactController *artifact.Controller,
	cronController *cron.Controller,
	statisticsController *statistics.Controller,
) *router.Controller {
	return &router.Controller{
		Space:      spaceController,
		Branch:     branchController,
		Tag:        tagController,
		Commit:     commitController,
		Repo:       repoController,
		User:       userController,
		Git:        gitController,
		Scan:       scanController,
		Pipeline:   pipelineController,
		Execution:  executionController,
		Runner:     runnerController,
		Svn:        svnController,
		Fit:        fitController,
		Webhook:    webhookController,
		Artifact:   artifactController,
		Cron:       cronController,
		Statistics: statisticsController,
	}
}
