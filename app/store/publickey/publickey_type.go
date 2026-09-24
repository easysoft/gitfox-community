// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package publickey

import (
	"database/sql"

	"github.com/easysoft/gitfox-community/app/model/publickey"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

type PublicKey struct {
	ID           int64           `db:"id" gorm:"column:id;primaryKey"`
	Account      string          `db:"account" gorm:"column:account"`
	VerifiedDate *datetime.Local `db:"verifiedDate" gorm:"column:verifiedDate"` // 公钥验证时间（若已验证），用于记录验证成功的时间。
	Identifier   string          `db:"identifier" gorm:"column:identifier"`
	Fingerprint  string          `db:"fingerprint" gorm:"column:fingerprint"` // 根据公钥内容计算的指纹（稳定摘要），用于快速查重与匹配。
	Content      sql.NullString  `db:"content" gorm:"column:content"`         // 公钥原始内容（如 ssh-ed25519 AAAA... user@host ），用于精确匹配与验证
	Comment      string          `db:"comment" gorm:"column:comment"`         // 公钥末尾的注释部分（通常是 user@host ），便于展示和识别。
	Type         string          `db:"type" gorm:"column:type"`               // 公钥算法类型（如 ssh-ed25519 、 ssh-rsa ），用于展示与后续策略控制。

	CreatedBy   string          `gorm:"column:createdBy"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedBy    string          `gorm:"column:editedBy"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
}

func (pk *PublicKey) TableName() string {
	return "ops_public_keys"
}

func (pk *PublicKey) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	pk.CreatedDate = now
	pk.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (pk *PublicKey) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	pk.EditedDate = now
	return nil
}

func (pk *PublicKey) toModel() publickey.PublicKey {
	return publickey.PublicKey{
		ID:           pk.ID,
		Account:      pk.Account,
		VerifiedDate: pk.VerifiedDate,
		Identifier:   pk.Identifier,
		Fingerprint:  pk.Fingerprint,
		Content:      pk.Content.String,
		Comment:      pk.Comment,
		Type:         pk.Type,
		CreatedBy:    pk.CreatedBy,
		CreatedDate:  pk.CreatedDate,
		EditedBy:     pk.EditedBy,
		EditedDate:   pk.EditedDate,
	}
}
