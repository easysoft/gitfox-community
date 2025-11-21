package token

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
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

func (s *tokensStore) CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	issuedAt := time.Now()

	var expiresAt time.Time
	if lifetime != nil {
		expiresAt = issuedAt.Add(*lifetime)
	}
	token := &Token{
		Account:    user.Account,
		Type:       string(constant.TokenTypePAT),
		Identifier: tokenIdentifier,
		ExpiresAt:  &expiresAt,
	}
	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create user pat token for user %d", user.ID)
	}

	return token.ToAuthToken(), nil
}

func (s *tokensStore) CreateAppPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	issuedAt := time.Now()

	var expiresAt time.Time
	if lifetime != nil {
		expiresAt = issuedAt.Add(*lifetime)
	}
	token := &Token{
		Account:    user.Account,
		Type:       string(constant.TokenTypeApp),
		Identifier: tokenIdentifier,
		ExpiresAt:  &expiresAt,
	}
	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create user pat token for user %d", user.ID)
	}

	return token.ToAuthToken(), nil
}
