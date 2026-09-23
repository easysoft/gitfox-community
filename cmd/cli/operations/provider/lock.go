// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/service/locker"
	"github.com/easysoft/gitfox-community/pkg/lock"
)

var LockWireSet = wire.NewSet(
	lock.WireSet,
	locker.WireSet,
)
