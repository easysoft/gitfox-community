package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/test"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	User     *user.Controller
	Commit   *commit.Controller
	Tag      *tag.Controller
	Branch   *branch.Controller
	Space    *space.SpaceController
	Git      *git.GitController
	Repo     *repo.Controller
	Scan     *scan.ScanController
	Pipeline *pipeline.Controller
	Test     *test.TestController
}
