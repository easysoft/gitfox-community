package model

import "github.com/easysoft/gitfox/types/enum"

type User struct {
	ID          int64  `db:"id"             gorm:"column:id;primaryKey" json:"id"`
	UID         string `db:"uid"            gorm:"column:uid"           json:"uid"`
	Email       string `db:"email"          gorm:"column:email"         json:"email"`
	DisplayName string `db:"display_name"   gorm:"column:display_name"  json:"display_name"`
	Admin       bool   `db:"admin"          gorm:"column:admin"         json:"admin"`
	Blocked     bool   `db:"blocked"        gorm:"column:blocked"       json:"blocked"`
	Salt        string `db:"salt"           gorm:"column:salt"          json:"-"`
	Created     int64  `db:"created"        gorm:"column:created"       json:"created"`
	Updated     int64  `db:"updated"        gorm:"column:updated"       json:"updated"`
	DeletedAt   int64  `db:"deleted_at"     gorm:"column:deleted_at"    json:"deleted_at,omitempty"`

	// User specific fields
	Password   string `db:"user_password"  gorm:"column:user_password"  json:"-"`
	Token      string `db:"user_token"     gorm:"column:user_token"     json:"-"`
	Expiration int64  `db:"user_expiration" gorm:"column:user_expiration" json:"-"`

	// Source is the source of the user account.
	Source enum.PrincipalSource `db:"user_source" gorm:"column:user_source" json:"-"`
}

// Principal represents the identity of an acting entity (User, ServiceAccount, Service).
type Principal struct {
	// TODO: int64 ID doesn't match DB
	ID          int64              `db:"principal_id"           gorm:"column:principal_id;primaryKey" json:"-"`
	UID         string             `db:"principal_uid"          gorm:"column:principal_uid"           json:"uid"`
	Email       string             `db:"principal_email"        gorm:"column:principal_email"         json:"email"`
	Type        enum.PrincipalType `db:"principal_type"         gorm:"column:principal_type"          json:"type"`
	DisplayName string             `db:"principal_display_name" gorm:"column:principal_display_name"  json:"display_name"`
	Admin       bool               `db:"principal_admin"        gorm:"column:principal_admin"         json:"admin"`

	// Should be part of principal or not?
	Blocked bool   `db:"principal_blocked"   gorm:"column:principal_blocked"  json:"blocked"`
	Salt    string `db:"principal_salt"      gorm:"column:principal_salt"     json:"-"`

	// Other info
	Created   int64 `db:"principal_created"    gorm:"column:principal_created"  json:"created"`
	Updated   int64 `db:"principal_updated"    gorm:"column:principal_updated"  json:"updated"`
	DeletedAt int64 `db:"principal_deleted_at"     gorm:"column:principal_deleted_at"    json:"deleted_at,omitempty"`

	// Source is the source of the user account.
	Source enum.PrincipalSource `db:"principal_user_source" gorm:"column:principal_user_source" json:"-"`
}

type PrincipalInfo struct {
	ID          int64              `json:"id"`
	UID         string             `json:"uid"`
	Password    string             `json:"-"`
	Token       string             `json:"-"`
	Expiration  int64              `json:"-"`
	DisplayName string             `json:"display_name"`
	Email       string             `json:"email"`
	Type        enum.PrincipalType `json:"type"`
	Created     int64              `json:"created"`
	Updated     int64              `json:"updated"`
}

func (p *PrincipalInfo) Identifier() int64 {
	return p.ID
}
