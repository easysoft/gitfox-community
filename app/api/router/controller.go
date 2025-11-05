package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/test"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	Branch *branch.BranchController
	Space  *space.SpaceController
	Test   *test.TestController
}
