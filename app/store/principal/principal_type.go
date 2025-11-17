package principal

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Principal struct {
	ID           int64  `db:"id" gorm:"column:id;primaryKey"`
	UID          string `db:"uid" gorm:"column:uid"`
	Email        string `db:"email" gorm:"column:email"`
	Admin        bool   `db:"admin" gorm:"column:admin"`
	Salt         string `db:"salt" gorm:"column:salt"`
	Created      int64  `db:"created" gorm:"column:created"`
	Updated      int64  `db:"updated" gorm:"column:updated"`
	UserPassword string `db:"user_password" gorm:"column:user_password"`

	CreatedBy   string     `gorm:"column:created_by"`
	CreatedDate *time.Time `gorm:"column:created_date"`
	EditedBy    string     `gorm:"column:edited_by"`
	EditedDate  *time.Time `gorm:"column:edited_date"`
	Deleted     bool       `gorm:"column:deleted"`
}

func (p *Principal) TableName() string {
	return "principals"
}

func (p *Principal) toUser() *principal.User {
	return &principal.User{
		ID:      p.ID,
		UID:     p.UID,
		Email:   p.Email,
		Salt:    p.Salt,
		IsAdmin: p.Admin,
	}
}
