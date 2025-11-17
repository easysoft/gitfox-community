package user

import (
	"context"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	principalStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	authPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userStore  principalStore.Interface
	tokenStore token.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 UserSrv 实例
func NewService(userStore principalStore.Interface, tokenStore token.Interface) *Service {
	return &Service{
		userStore:  userStore,
		tokenStore: tokenStore,
	}
}

func (s *Service) FindUserByUID(ctx context.Context, uid string) (*principal.User, error) {
	return s.userStore.FindUserByUID(ctx, uid)
}

func (s *Service) FindUserByEmail(ctx context.Context, email string) (*principal.User, error) {
	return s.userStore.FindUserByEmail(ctx, email)
}

func (s *Service) CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func (s *Service) IssueToken(ctx context.Context, user *principal.User) (*auth.TokenResponse, error) {
	tokenIdentifier, err := GenerateSessionTokenIdentifier()
	if err != nil {
		return nil, err
	}

	token, err := s.tokenStore.CreateUserSession(ctx, user, tokenIdentifier)
	if err != nil {
		return nil, err
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForToken(token, user.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt token: %w", err)
	}

	return &auth.TokenResponse{Token: *token, AccessToken: jwtToken}, nil
}

func (s *Service) IssuePAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error) {
	token, err := s.tokenStore.CreateUserPAT(ctx, user, identifier, lifetime)
	if err != nil {
		return nil, err
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForToken(token, user.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt token: %w", err)
	}

	return &auth.TokenResponse{Token: *token, AccessToken: jwtToken}, nil
}
