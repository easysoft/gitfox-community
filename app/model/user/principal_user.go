// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package principal

// 用户实体
type User struct {
	ID          int64  `json:"id"`
	Account     string `json:"account"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"-"`
	IsAdmin     bool   `json:"is_admin"`

	// todo: 后续根据需要添加其他字段
}

// Identifier returns the unique identifier for the user (account).
// Identifier 返回用户的唯一标识符（账号）。
func (u *User) Identifier() string {
	return u.Account
}

// Principal represents the identity of an acting entity ( ServiceAccount, Service).
// Principal 主要代表一个活动实体（服务帐户、服务）的标识。
type Principal struct {
	App         string `json:"app"`
	DisplayName string `json:"display_name"`
	Admin       bool   `json:"admin"`
	Email       string `json:"email"`
	// todo: 后续根据需要添加其他字段
}
