// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package provider

import (
	publickeySvc "github.com/easysoft/gitfox-community/app/service/publickey"
	publickeyStore "github.com/easysoft/gitfox-community/app/store/publickey"
	"github.com/google/wire"
)

var PublicKeyWireSet = wire.NewSet(
	publickeyStore.NewPublicKeyStore,
	publickeySvc.NewPublicKeyService,
)
