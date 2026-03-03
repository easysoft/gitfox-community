package publickey

import (
	"context"
	"time"

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

// ListByIdentifier 根据标识符查询公钥列表
func (s *Store) ListByIdentifier(ctx context.Context, identifier string, account string) ([]publickey.PublicKey, error) {
	var publicKeys []PublicKey
	if err := s.db.WithContext(ctx).Where("identifier = ? and account = ?", identifier, account).Find(&publicKeys).Error; err != nil {
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

		// 校验标识符是否已存在
		existingKeys, err := s.ListByIdentifier(ctx, publicKey.Identifier, publicKey.Account)
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		if len(existingKeys) > 0 {
			err = errors.New("Identifier is already in use")
			err = errors.WithI18nKey(err, i18n.AuthPubKeyAddIdentifierInUse)
			return err
		}

		// 先查询是否存在相同指纹的公钥
		existingKeys, err = s.ListByFingerprint(ctx, publicKey.Fingerprint)
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
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

func (s *Store) UpdateByID(ctx context.Context, account string, id int64, publicKey PublicKey) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if publicKey.Identifier != "" {
			// 校验标识符是否已存在
			existingKeys, err := s.ListByIdentifier(ctx, publicKey.Identifier, account)
			if err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			}

			if len(existingKeys) > 0 {
				err = errors.New("Identifier is already in use")
				err = errors.WithI18nKey(err, i18n.AuthPubKeyAddIdentifierInUse)
				return err
			}
		}

		if publicKey.Fingerprint != "" {
			// 校验是否已存在相同内容的公钥
			existingKeys, err := s.ListByFingerprint(ctx, publicKey.Fingerprint)
			if err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			}
			for _, existingKey := range existingKeys {
				if existingKey.Content == publicKey.Content {
					err = errors.New("Key is already in use")
					err = errors.WithI18nKey(err, i18n.AuthPubKeyAddAlreadyInUse)
					return err
				}
			}
		}

		result := tx.Model(&PublicKey{}).Where("account = ? and id = ?", account, id).Updates(&publicKey)
		if result.Error != nil {
			return database.ProcessSQLErrorf(ctx, result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.NotFound.Newf("public key id %d not found", id)
		}
		return nil
	})
}

func (s *Store) DeleteByID(ctx context.Context, account string, id int64) error {
	result := s.db.WithContext(ctx).
		Where("account = ? and id = ?", account, id).
		Delete(&PublicKey{})
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound.Newf("public key id %d not found", id)
	}
	return nil
}

// MarkAsVerified updates the public key to mark it as verified.
// MarkAsVerified更新公钥标记为已验证。
func (s *Store) MarkAsVerified(ctx context.Context, id int64) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&PublicKey{}).
		Where("id = ?", id).
		Updates(&PublicKey{VerifiedDate: &now}).Error
}

func (s *Store) ListByAccount(ctx context.Context, account string, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error) {
	tx := s.db.WithContext(ctx).Model(&PublicKey{}).Where("account = ?", account)

	// 按 创建时间倒序排序，确保最新的公钥排在前面
	tx.Order("createdDate DESC")

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
