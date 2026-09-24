// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
)

func newTestSession(account string) *authModel.Session {
	return &authModel.Session{User: &principal.User{Account: account}}
}
