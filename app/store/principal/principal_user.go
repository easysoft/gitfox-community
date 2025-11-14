package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

func NewPrincipalStore(db *gorm.DB) Interface {
	return &Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

func (s *Store) FindUserByID(ctx context.Context, id int64) (*principal.User, error) {
	var principal Principal
	principal.ID = id
	if err := s.db.WithContext(ctx).First(&principal).Error; err != nil {
		return nil, err
	}

	return principal.toUser(), nil
}

func (s *Store) FindUserByUID(ctx context.Context, uid string) (*principal.User, error) {
	var principal Principal
	err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&principal).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.DBError.Wrapf(err, "uid: %s", uid)
		err = errors.WithI18nKey(err, i18n.DbNotfound)
		return nil, err
	}
	if err != nil {
		return nil, errors.DBError.Wrapf(err, "uid: %s", uid)
	}
	return principal.toUser(), nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*principal.User, error) {
	var principal Principal
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&principal).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.DBError.Wrapf(err, "email: %s", email)
		err = errors.WithI18nKey(err, i18n.DbNotfound)
		return nil, err
	}
	if err != nil {
		return nil, errors.DBError.Wrapf(err, "email: %s", email)
	}
	return principal.toUser(), nil
}
