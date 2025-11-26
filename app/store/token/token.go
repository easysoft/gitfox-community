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
	if err := s.db.WithContext(ctx).Where("id = ?", id).Find(&out).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get token by id %d", id)
	}
	if out.ID == 0 {
		return nil, errors.NotFound.Newf("token not found by id %d", id)
	}

	return &out, nil
}

func (s *tokensStore) CreateUserPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	token := &Token{
		Account:    user.Account,
		Type:       string(constant.TokenTypePAT),
		Identifier: tokenIdentifier,
	}

	// 如果提供了 lifetime，则设置过期时间
	if lifetime != nil {
		expirationDate := time.Now().Add(*lifetime)
		token.ExpirationDate = &expirationDate
	}
	// 否则 ExpirationDate 保持为 nil，表示永不过期
	// CreatedDate 会由 BeforeCreate 钩子自动设置

	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create user pat token for user %d", user.ID)
	}

	return token.ToAuthToken(), nil
}

func (s *tokensStore) CreateAppPAT(ctx context.Context, user *principal.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	token := &Token{
		Account:    user.Account,
		Type:       string(constant.TokenTypeApp),
		Identifier: tokenIdentifier,
	}

	// 如果提供了 lifetime，则设置过期时间
	if lifetime != nil {
		expirationDate := time.Now().Add(*lifetime)
		token.ExpirationDate = &expirationDate
	}

	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "create app pat token for user %s", user.Account)
	}

	return token.ToAuthToken(), nil
}

func (s *tokensStore) UpdateToken(ctx context.Context, id int64, token string) error {
	return s.db.WithContext(ctx).Model(&Token{}).Where("id = ?", id).Update("token", token).Error
}

func (s *tokensStore) GetByIdentifier(ctx context.Context, account string, identifier string) (*Token, error) {
	var token Token

	if err := s.db.WithContext(ctx).Where("account = ? AND identifier = ?", account, identifier).Find(&token).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get token by account %s and identifier %s", account, identifier)
	}

	if token.ID == 0 {
		return nil, errors.NotFound.Newf("token not found by account %s and identifier %s", account, identifier)
	}

	return &token, nil
}

func (s *tokensStore) GetDevopsToken(ctx context.Context) (*Entry, error) {
	var entry Entry

	if err := s.db.WithContext(ctx).Where("name = 'GitFox' and code = 'gitfox' and deleted = 0").Find(&entry).Error; err != nil {
		return nil, errors.DBError.Wrapf(err, "get entry by code gitfox")
	}

	if entry.ID == 0 {
		return nil, errors.NotFound.Newf("entry not found by code gitfox")
	}

	return &entry, nil
}
