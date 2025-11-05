package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	Space  *space.SpaceController
	Branch *branch.BranchController
}
