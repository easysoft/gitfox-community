package user

import (
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

type User struct {
	ID       int64  `db:"id" gorm:"column:id;primaryKey"`
	Account  string `db:"account" gorm:"column:account"`
	Password string `db:"password" gorm:"column:password"`
	Role     string `db:"role" gorm:"column:role"`
	Email    string `db:"email" gorm:"column:email"`
}

func (p *User) TableName() string {
	return "zt_user"
}

func (p *User) toUser() *principal.User {
	return &principal.User{
		ID:       p.ID,
		Account:  p.Account,
		Email:    p.Email,
		Password: p.Password,
		IsAdmin:  p.Role == "admin",
	}
}
