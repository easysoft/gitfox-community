// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/space"
	spaceSvc "github.com/easysoft/gitfox-community/app/service/space"
	spaceStore "github.com/easysoft/gitfox-community/app/store/space"
)

var SpaceWireSet = wire.NewSet(
	space.NewSpaceController,
	wire.Bind(new(spaceSvc.Interface), new(*spaceSvc.Service)),
	// 提供创建实现类实例的方法（关键修复）
	spaceSvc.NewSpaceSvc,
	wire.Bind(new(spaceStore.Interface), new(*spaceStore.SpaceStore)),
	spaceStore.NewSpaceStore,
)
