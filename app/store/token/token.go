package token

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/gotidy/ptr"

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

func (s *tokensStore) CreateUserSession(ctx context.Context, user *principal.User, tokenIdentifier string) (*auth.Token, error) {
	token := &Token{
		PrincipalID: user.ID,
		Type:        string(constant.TokenTypeSession),
		Identifier:  tokenIdentifier,
	}
	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create user session token for user %d", user.ID)
	}

	return token.ToAuthToken(), nil
}

func (s *tokensStore) CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	issuedAt := time.Now()

	var expiresAt *int64
	if lifetime != nil {
		expiresAt = ptr.Int64(issuedAt.Add(*lifetime).UnixMilli())
	}
	token := &Token{
		PrincipalID: user.ID,
		Type:        string(constant.TokenTypePAT),
		Identifier:  tokenIdentifier,
		ExpiresAt:   expiresAt,
	}
	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create user pat token for user %d", user.ID)
	}

	return token.ToAuthToken(), nil
}
