package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

func NewPublicKeyStore(db *gorm.DB) Interface {
	return &Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

// ListByFingerprint 根据指纹查询公钥列表
func (s *Store) ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error) {
	var publicKeys []PublicKey
	if err := s.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).Find(&publicKeys).Error; err != nil {
		return nil, err
	}

	result := make([]publickey.PublicKey, len(publicKeys))
	for i := range publicKeys {
		result[i] = publicKeys[i].toModel()
	}

	return result, nil
}

// MarkAsVerified updates the public key to mark it as verified.
func (s *Store) MarkAsVerified(ctx context.Context, id int64, verified int64) error {

	if err := s.db.WithContext(ctx).
		Model(&PublicKey{}).
		Where("id = ?", id).
		Update("verified", verified).Error; err != nil {
		return err
	}

	return nil
}
func (s *Store) Create(ctx context.Context, publicKey PublicKey) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 先查询是否存在相同指纹的公钥
		existingKeys, err := s.ListByFingerprint(ctx, publicKey.Fingerprint)
		if err != nil {
			err = errors.Wrap(err, "failed to read keys by fingerprint")
			err = errors.WithI18nKey(err, i18n.AuthPubKeyNotFoundByFingerprint)
			return err
		}

		// 校验是否已存在相同内容的公钥
		for _, existingKey := range existingKeys {
			if existingKey.Content == publicKey.Content {
				err = errors.New("Key is already in use")
				err = errors.WithI18nKey(err, i18n.AuthPubKeyAlreadyInUse)
				return err
			}
		}

		// 创建公钥
		if err := tx.Create(&publicKey).Error; err != nil {
			err = errors.Wrap(err, "failed to insert public key")
			err = errors.WithI18nKey(err, i18n.AuthPubKeyCreateError)
			return err
		}

		return nil
	})
}
