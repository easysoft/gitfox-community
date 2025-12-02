package user

import (
	"context"

	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

func NewUserStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

func (s *Store) FindByID(ctx context.Context, id int64) (*principal.User, error) {
	var user User
	user.ID = id
	if err := s.db.WithContext(ctx).First(&user).Error; err != nil {
		return nil, err
	}

	if user.ID == 0 {
		return nil, errors.WithI18nKey(errors.NotFound.Newf("id %d", id), i18n.DbNotfound)
	}

	return user.toUser(), nil
}

func (s *Store) FindByAccount(ctx context.Context, account string) (*principal.User, error) {
	var user User
	err := s.db.WithContext(ctx).Where("account = ?", account).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.DBError.Wrapf(err, "account: %s", account)
		err = errors.WithI18nKey(err, i18n.DbNotfound)
		return nil, err
	}
	if err != nil {
		return nil, errors.DBError.Wrapf(err, "account: %s", account)
	}

	if user.ID == 0 {
		return nil, errors.WithI18nKey(errors.NotFound.Newf("account %s", account), i18n.DbNotfound)
	}

	return user.toUser(), nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (*principal.User, error) {
	var user User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.DBError.Wrapf(err, "email: %s", email)
		err = errors.WithI18nKey(err, i18n.DbNotfound)
		return nil, err
	}
	if err != nil {
		return nil, errors.DBError.Wrapf(err, "email: %s", email)
	}

	if user.ID == 0 {
		return nil, errors.WithI18nKey(errors.NotFound.Newf("email %s", email), i18n.DbNotfound)
	}

	return user.toUser(), nil
}
