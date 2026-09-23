// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package router

import (
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
	"github.com/easysoft/gitfox-community/app/api/controller/test"
	"github.com/easysoft/gitfox-community/app/api/controller/user"
	"github.com/easysoft/gitfox-community/app/api/controller/webhook"
)

// Controller 聚合所有的 HTTP 处理器
type Controller struct {
	User       *user.Controller
	Commit     *commit.Controller
	Tag        *tag.Controller
	Branch     *branch.Controller
	Space      *space.SpaceController
	Svn        *svn.Controller
	Webhook    *webhook.Controller
	Git        *git.GitController
	Repo       *repo.Controller
	Scan       *scan.ScanController
	Pipeline   *pipeline.Controller
	Runner     *runner.Controller
	Execution  *execution.Controller
	Test       *test.TestController
	Fit        *fit.Controller
	Cron       *cron.Controller
	Artifact   *artifact.Controller
	Statistics *statistics.Controller
}
