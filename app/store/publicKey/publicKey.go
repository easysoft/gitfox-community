package publicKey

import (
	"context"

	publickey "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publicKey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"github.com/Masterminds/squirrel"

	"gorm.io/gorm"
)

var _ store.IPublicKeyStore = (*PublicKeyStore)(nil)

func NewPublicKeyStore(db *gorm.DB) store.IPublicKeyStore {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	return &PublicKeyStore{
		db: db,
		sb: factory.CreateQueryBuilder(),
	}
}

type PublicKeyStore struct {
	db *gorm.DB
	sb sqlbuilder.QueryBuilder
}

// ListByFingerprint 根据指纹查询公钥列表
func (s *PublicKeyStore) ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error) {
	// 使用 Builder 生成 SQL 语句
	sql, args, err := s.ListByFingerprintBuilder(ctx, fingerprint).ToSql()
	if err != nil {
		return nil, err
	}

	// 执行查询
	var publicKeys []publickey.PublicKey
	err = s.db.WithContext(ctx).Raw(sql, args...).Scan(&publicKeys).Error
	if err != nil {
		return nil, err
	}
	return publicKeys, nil
}

// ListByFingerprintBuilder 构建根据指纹查询的SQL
func (s *PublicKeyStore) ListByFingerprintBuilder(ctx context.Context, fingerprint string) sqlbuilder.SelectBuilder {
	return s.sb.
		Select("*").
		From("public_keys").
		Where(squirrel.Eq{"public_key_fingerprint": fingerprint})
}

// MarkAsVerified updates the public key to mark it as verified.
func (s PublicKeyStore) MarkAsVerified(ctx context.Context, id int64, verified int64) error {
	// if err := s.db.Model(&PublicKey{}).Where("id = ?", id).Update("verified", verified).Error; err != nil {
	// 	return database.ProcessGormSQLErrorf(ctx, err, "Failed to mark public key as verified")
	// }

	return nil
}
