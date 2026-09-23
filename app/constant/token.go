// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

// TokenType represents the type of the JWT token.
type TokenType string

const (
	// TokenTypePAT is a personal access token.
	TokenTypePAT TokenType = "pat"

	TokenTypeApp TokenType = "app"
)
