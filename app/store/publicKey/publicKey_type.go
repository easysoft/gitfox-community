package publicKey

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

type PublicKey struct {
	ID          int64               `db:"id" gorm:"column:id;primaryKey"`
	PrincipalID int64               `db:"principal_id" gorm:"column:principal_id"`
	Created     int64               `db:"created" gorm:"column:created"`
	Verified    *int64              `db:"verified" gorm:"column:verified"`
	Identifier  string              `db:"identifier" gorm:"column:identifier"`
	Usage       constant.PublicKeyUsage `db:"usage" gorm:"column:usage"`
	Fingerprint string              `db:"fingerprint" gorm:"column:fingerprint"`
	Content     string              `db:"content" gorm:"column:content"`
	Comment     string              `db:"comment" gorm:"column:comment"`
	Type        string              `db:"type" gorm:"column:type"`
}

func (pk *PublicKey) TableName() string {
	return "public_keys"
}
