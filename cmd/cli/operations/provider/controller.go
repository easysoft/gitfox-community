package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
)

var ControllersWireSet = wire.NewSet(
	NewControllers,
)

// NewControllers 创建所有 HTTP 处理器的聚合实例
func NewControllers(
	spaceController *space.SpaceController,
	branchController *branch.Controller,
	tagController *tag.Controller,
	commitController *commit.Controller,
) *router.Controller {
	return &router.Controller{
		Space:  spaceController,
		Branch: branchController,
		Tag:    tagController,
		Commit: commitController,
	}
}
