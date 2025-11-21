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

	ExpiresDate *time.Time `gorm:"column:expires_date"` // ExpiresDate is an optional date that if specified restricts the validity of a token.
	CreatedBy   string     `gorm:"column:created_by"`
	CreatedDate *time.Time `gorm:"column:created_date"`
	EditedBy    string     `gorm:"column:edited_by"`
	EditedDate  *time.Time `gorm:"column:edited_date"`
	Deleted     int8       `gorm:"column:deleted"`
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
		ID:          t.ID,
		Account:     t.Account,
		Type:        constant.TokenType(t.Type),
		Identifier:  t.Identifier,
		ExpiresDate: t.ExpiresDate,
		CreatedDate: t.CreatedDate,
	}
}
