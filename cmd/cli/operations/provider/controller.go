package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
)

var ControllersWireSet = wire.NewSet(
	NewControllers,
)

// NewControllers 创建所有 HTTP 处理器的聚合实例
func NewControllers(spaceController *space.SpaceController) *router.Controller {
	return &router.Controller{
		Space: spaceController,
	}
}
