package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gorm.io/gorm"
)

var _ IPublicKeyStore = (*PublicKeyStore)(nil)

func NewPublicKeyStore(db *gorm.DB) IPublicKeyStore {
	return &PublicKeyStore{
		db: db,
	}
}

type PublicKeyStore struct {
	db *gorm.DB
}

// ListByFingerprint 根据指纹查询公钥列表
func (s *PublicKeyStore) ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error) {
	// 使用 gorm 查询
	var publicKeys []PublicKey
	var query PublicKey
	query.Fingerprint = fingerprint
	if err := s.db.WithContext(ctx).Where(&query).Find(&publicKeys).Error; err != nil {
		return nil, err
	}

	// 这里暂时保持返回类型,转换逻辑应在 service 层完成
	result := make([]publickey.PublicKey, len(publicKeys))
	for i := range publicKeys {
		// TODO: 应该在 service 层完成转换
		result[i] = publickey.PublicKey(publicKeys[i])
	}

	return result, nil
}

// MarkAsVerified updates the public key to mark it as verified.
func (s *PublicKeyStore) MarkAsVerified(ctx context.Context, id int64, verified int64) error {
	// 使用 gorm 更新 verified 字段
	var query PublicKey
	query.ID = id
	if err := s.db.WithContext(ctx).
		Model(&PublicKey{}).
		Where(&query).
		Update("verified", verified).Error; err != nil {
		return err
	}

	return nil
}
