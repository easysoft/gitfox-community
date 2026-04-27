package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/fit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/test"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/webhook"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	User      *user.Controller
	Commit    *commit.Controller
	Tag       *tag.Controller
	Branch    *branch.Controller
	Space     *space.SpaceController
	Webhook   *webhook.Controller
	Git       *git.GitController
	Repo      *repo.Controller
	Scan      *scan.ScanController
	Pipeline  *pipeline.Controller
	Runner    *runner.Controller
	Execution *execution.Controller
	Test      *test.TestController
	Fit       *fit.Controller
	Artifact  *artifact.Controller
}
