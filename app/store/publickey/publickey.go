package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
				err = errors.WithI18nKey(err, i18n.AuthPubKeyAddAlreadyInUse)
				return err
			}
		}

		// 创建公钥
		if err := tx.Create(&publicKey).Error; err != nil {
			err = errors.Wrap(err, "failed to insert public key")
			err = errors.WithI18nKey(err, i18n.AuthPubKeyAddError)
			return err
		}

		return nil
	})
}

// MarkAsVerified updates the public key to mark it as verified.
// MarkAsVerified更新公钥标记为已验证。
func (s *Store) MarkAsVerified(ctx context.Context, id int64, verified int64) error {

	if err := s.db.WithContext(ctx).Model(&PublicKey{}).Where("id = ?", id).Update("verified", verified).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) ListByAccount(ctx context.Context, account string, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error) {
	tx := s.db.WithContext(ctx).Model(&PublicKey{}).Where("account = ? AND deleted = ?", account, 0)

	var keys []*PublicKey
	pageResult, err := database.PaginateDefault(ctx, tx, params.Page, params.PageSize, &keys)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// Convert store models to API models
	result := make([]*publickey.PublicKey, len(keys))
	for i := range keys {
		model := keys[i].toModel()
		result[i] = &model
	}

	return &database.PageResult[*publickey.PublicKey]{
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
		Data:     result,
	}, nil
}
