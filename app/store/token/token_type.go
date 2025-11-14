package token

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
)

type Token struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	PrincipalID int64  `gorm:"column:principal_id"`
	Type        string `gorm:"column:type"`
	Identifier  string `gorm:"column:identifier"`
	// ExpiresAt is an optional unix time that if specified restricts the validity of a token.
	ExpiresAt *int64 `gorm:"column:expires_at"`
	// IssuedAt is the unix time at which the token was issued.
	CreatedBy   string     `gorm:"column:created_by"`
	CreatedDate *time.Time `gorm:"column:created_date"`
	EditedBy    string     `gorm:"column:edited_by"`
	EditedDate  *time.Time `gorm:"column:edited_date"`
	Deleted     bool       `gorm:"column:deleted"`
}

func (Token) TableName() string {
	return "tokens"
}

func (t *Token) ToAuthToken() *auth.Token {
	return &auth.Token{
		ID:          t.ID,
		PrincipalID: t.PrincipalID,
		Type:        constant.TokenType(t.Type),
		Identifier:  t.Identifier,
		ExpiresAt:   t.ExpiresAt,
	}
}
