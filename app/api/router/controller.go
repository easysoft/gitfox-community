package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/test"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	Space *space.SpaceController
	Test  *test.TestController
}
