// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package token

import (
	"database/sql"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	"gorm.io/gorm"
)

type Token struct {
	ID         int64          `gorm:"column:id;primaryKey"`
	Account    string         `gorm:"column:account"`
	Type       string         `gorm:"column:type"`
	Identifier string         `gorm:"column:identifier"`
	Token      sql.NullString `gorm:"column:token"`

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

type Entry struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string     `gorm:"column:name;type:varchar(50);not null;default:''"`
	Account     string     `gorm:"column:account;type:varchar(30);not null;default:''"`
	Code        string     `gorm:"column:code;type:varchar(20);not null;default:''"`
	Key         string     `gorm:"column:key;type:varchar(32);not null;default:''"`
	FreePasswd  string     `gorm:"column:freePasswd;type:enum('0','1');not null;default:'0'"`
	IP          string     `gorm:"column:ip;type:varchar(100);not null;default:''"`
	Desc        string     `gorm:"column:desc;type:mediumtext"`
	CreatedBy   string     `gorm:"column:createdBy;type:varchar(30);not null;default:''"`
	CreatedDate *time.Time `gorm:"column:createdDate"`
	CalledTime  uint       `gorm:"column:calledTime;not null;default:0"`
	EditedBy    string     `gorm:"column:editedBy;type:varchar(30);not null;default:''"`
	EditedDate  *time.Time `gorm:"column:editedDate"`
	Deleted     uint8      `gorm:"column:deleted;not null;default:0"`
}

func (Entry) TableName() string {
	return "zt_entry"
}

func (e *Entry) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	e.CreatedDate = &now
	e.EditedDate = &now
	return nil
}

func (e *Entry) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	e.EditedDate = &now
	return nil
}
