// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/fit"
	fitSvc "github.com/easysoft/gitfox-community/app/service/fit"
)

var FitWireSet = wire.NewSet(
	fit.NewController,
	wire.Bind(new(fitSvc.Interface), new(*fitSvc.Service)),
	fitSvc.NewService,
)
