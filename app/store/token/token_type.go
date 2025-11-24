package token

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gorm.io/gorm"
)

type Token struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	Account    string `gorm:"column:account"`
	Type       string `gorm:"column:type"`
	Identifier string `gorm:"column:identifier"`
	Token      string `gorm:"column:token"`

	ExpirationDate *time.Time `gorm:"column:expirationDate"` // ExpiresDate is an optional date that if specified restricts the validity of a token.
	CreatedBy      string     `gorm:"column:createdBy"`
	CreatedDate    *time.Time `gorm:"column:createdDate"`
	EditedBy       string     `gorm:"column:editedBy"`
	EditedDate     *time.Time `gorm:"column:editedDate"`
}

func (Token) TableName() string {
	return "ops_tokens"
}

func (r *Token) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedDate = &now
	r.EditedDate = &now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *Token) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	r.EditedDate = &now
	return nil
}

func (t *Token) ToAuthToken() *auth.Token {
	return &auth.Token{
		ID:             t.ID,
		Account:        t.Account,
		Type:           constant.TokenType(t.Type),
		Identifier:     t.Identifier,
		ExpirationDate: t.ExpirationDate,
		CreatedDate:    t.CreatedDate,
	}
}
