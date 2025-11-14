package principal

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Principal struct {
	ID           int64                    `db:"id" gorm:"column:id;primaryKey"`
	UID          string                   `db:"uid" gorm:"column:uid"`
	UIDUnique    string                   `db:"uid_unique" gorm:"column:uid_unique"`
	Email        string                   `db:"email" gorm:"column:email"`
	Type         constant.PrincipalType   `db:"type" gorm:"column:type"`
	DisplayName  string                   `db:"display_name" gorm:"column:display_name"`
	Admin        bool                     `db:"admin" gorm:"column:admin"`
	Blocked      bool                     `db:"blocked" gorm:"column:blocked"`
	Salt         string                   `db:"salt" gorm:"column:salt"`
	Created      int64                    `db:"created" gorm:"column:created"`
	Updated      int64                    `db:"updated" gorm:"column:updated"`
	UserPassword string                   `db:"user_password" gorm:"column:user_password"`
	SAParentType string                   `db:"sa_parent_type" gorm:"column:sa_parent_type"`
	SAParentID   int                      `db:"sa_parent_id" gorm:"column:sa_parent_id"`
	UserSource   constant.PrincipalSource `db:"user_source" gorm:"column:user_source"`
	DeletedAt    int64                    `db:"deleted_at" gorm:"column:deleted_at"`
}

func (p *Principal) TableName() string {
	return "principals"
}

func (p *Principal) toUser() *principal.User {
	return &principal.User{
		ID:          p.ID,
		UID:         p.UID,
		Email:       p.Email,
		DisplayName: p.DisplayName,
		Admin:       p.Admin,
		Blocked:     p.Blocked,
		Salt:        p.Salt,
		Created:     p.Created,
		Updated:     p.Updated,
		DeletedAt:   p.DeletedAt,
	}
}
