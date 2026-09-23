// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package auth

import principal "github.com/easysoft/gitfox-community/app/model/user"

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal *principal.Principal

	// User is the end principal. default to Principal, override if SudoUser exist
	User *principal.User `json:"user"`
}
