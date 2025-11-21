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
	if err := s.db.WithContext(ctx).Find(&out).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get token by id %d", id)
	}
	if out.ID == 0 {
		return nil, errors.NotFound.Newf("token not found by id %d", id)
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
		return nil, errors.DBError.Wrapf(err, "create app pat token for user %s", user.Account)
	}

	return token.ToAuthToken(), nil
}

func (s *tokensStore) UpdateToken(ctx context.Context, id int64, token string) error {
	return s.db.WithContext(ctx).Model(&Token{}).Where("id = ?", id).Update("token", token).Error
}

func (s *tokensStore) GetByIdentifier(ctx context.Context, identifier string) (*Token, error) {
	var out Token
	out.Identifier = identifier
	if err := s.db.WithContext(ctx).Find(&out).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get token by identifier %s", identifier)
	}

	if out.ID == 0 {
		return nil, errors.NotFound.Newf("token not found by identifier %s", identifier)
	}

	return &out, nil
}
