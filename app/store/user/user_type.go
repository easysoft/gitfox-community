// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package user

import (
	principal "github.com/easysoft/gitfox-community/app/model/user"
)

type User struct {
	ID       int64  `db:"id" gorm:"column:id;primaryKey"`
	Account  string `db:"account" gorm:"column:account"`
	Password string `db:"password" gorm:"column:password"`
	Role     string `db:"role" gorm:"column:role"`
	Email    string `db:"email" gorm:"column:email"`
	RealName string `db:"realname" gorm:"column:realname"`
}

func (p *User) TableName() string {
	return "zt_user"
}

func (p *User) toUser() *principal.User {
	return &principal.User{
		ID:          p.ID,
		Account:     p.Account,
		DisplayName: p.RealName,
		Email:       p.Email,
		Password:    p.Password,
		IsAdmin:     p.Role == "admin",
	}
}
