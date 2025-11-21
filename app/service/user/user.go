package user

import (
	"context"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	authPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userStore  user.Interface
	tokenStore token.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 UserSrv 实例
func NewService(userStore user.Interface, tokenStore token.Interface) *Service {
	return &Service{
		userStore:  userStore,
		tokenStore: tokenStore,
	}
}

func (s *Service) FindUserByID(ctx context.Context, id int64) (*principal.User, error) {
	return s.userStore.FindUserByID(ctx, id)
}

func (s *Service) FindUserByAccount(ctx context.Context, account string) (*principal.User, error) {
	return s.userStore.FindUserByAccount(ctx, account)
}

func (s *Service) FindUserByEmail(ctx context.Context, email string) (*principal.User, error) {
	return s.userStore.FindUserByEmail(ctx, email)
}

func (s *Service) CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func (s *Service) IssueUserPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error) {
	token, err := s.tokenStore.CreateUserPAT(ctx, user, identifier, lifetime)
	if err != nil {
		return nil, err
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForToken(token, user.Account)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt token: %w", err)
	}

	return &auth.TokenResponse{Token: *token, AccessToken: jwtToken}, nil
}

func (s *Service) IssueAppPAT(ctx context.Context, user *principal.User, identifier string, isAdmin bool, lifetime *time.Duration) (*auth.TokenResponse, error) {
	token, err := s.tokenStore.CreateAppPAT(ctx, user, identifier, lifetime)
	if err != nil {
		return nil, err
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForAppToken(token, user.Account, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt token: %w", err)
	}

	return &auth.TokenResponse{Token: *token, AccessToken: jwtToken}, nil
}
