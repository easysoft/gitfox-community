package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
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
	tagController *tag.Controller,
	commitController *commit.Controller,
	repoController *repo.Controller,
	userController *user.Controller,
	gitController *git.GitController,
	scanController *scan.ScanController,
) *router.Controller {
	return &router.Controller{
		Space:     spaceController,
		Branch:    branchController,
		Tag:       tagController,
		Commit:    commitController,
		Repo:      repoController,
		User:      userController,
		Git:       gitController,
		Scan:      scanController,
		Pipeline:  pipelineController,
		Execution: executionController,
		Runner:    runnerController,
	}
}
