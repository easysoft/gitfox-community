package user

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	authPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

func GenerateSessionTokenIdentifier() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return fmt.Sprintf("login-%d-%04d", time.Now().Unix(), r.Int64()), nil
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

func (s *Service) IssueAppPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error) {
	token, err := s.tokenStore.CreateAppPAT(ctx, user, identifier, lifetime)
	if err != nil {
		return nil, err
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForAppToken(token, user.Account, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt token: %w", err)
	}

	return &auth.TokenResponse{Token: *token, AccessToken: jwtToken}, nil
}

const devopsAdminAccount = "devops_admin"
const devopsAdminIdentifier = "devops_admin_identifier"

func (s *Service) IssueAdminPAT() (string, error) {
	ctx := context.Background()

	tokens, err := s.tokenStore.GetByIdentifier(ctx, devopsAdminIdentifier)
	if err != nil && !errors.IsNotFound(err) {
		return "", errors.Wrapf(err, "get devops admin token")
	}

	if tokens != nil {
		return tokens.Identifier, nil
	}

	// 没有的话生成下
	devopsAdmin := &principal.User{Account: devopsAdminAccount, IsAdmin: true}
	token, err := s.tokenStore.CreateAppPAT(ctx, devopsAdmin, devopsAdminIdentifier, nil)
	if err != nil {
		return "", errors.Wrapf(err, "create devops admin token for user %s", devopsAdmin.Account)
	}

	// create jwt token.
	jwtToken, err := authPkg.GenerateForAppToken(token, devopsAdmin.Account, true)
	if err != nil {
		return "", errors.Wrapf(err, "create devops admin jwt token for user %s", devopsAdmin.Account)
	}

	updateErr := s.tokenStore.UpdateToken(ctx, token.ID, jwtToken)
	if updateErr != nil {
		return "", errors.Wrapf(updateErr, "update devops admin token for user %s", devopsAdmin.Account)
	}

	return jwtToken, nil
}
