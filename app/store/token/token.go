package token

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"gorm.io/gorm"
)

type tokensStore struct {
	db *gorm.DB
}

func NewTokensStore(db *gorm.DB) *tokensStore {
	return &tokensStore{db: db}
}

func (s *tokensStore) GetByID(ctx context.Context, id int64) (*Token, error) {
	var out Token
	out.ID = id
	if err := s.db.WithContext(ctx).First(&out).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get token by id %d", id)
	}

	return &out, nil
}
