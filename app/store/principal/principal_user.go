package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gorm.io/gorm"
)

var _ IPrincipalStore = (*PrincipalStore)(nil)

func NewPrincipalStore(db *gorm.DB) IPrincipalStore {
	return &PrincipalStore{
		db: db,
	}
}

type PrincipalStore struct {
	db *gorm.DB
}

func (s *PrincipalStore) FindUserByID(ctx context.Context, id int64) (*principal.User, error) {
	var user Principal
	user.ID = id
	if err := s.db.WithContext(ctx).Where(&user).First(&user).Error; err != nil {
		return nil, err
	}

	return toUser(&user), nil
}

func (s *PrincipalStore) FindUserByUID(ctx context.Context, uid string) (*principal.User, error) {
	var user Principal
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error; err != nil {
		return nil, err
	}
	return toUser(&user), nil
}
