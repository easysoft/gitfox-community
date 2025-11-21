package user

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
)

type Interface interface {
	FindUserByUID(ctx context.Context, uid string) (*principal.User, error)
	FindUserByEmail(ctx context.Context, email string) (*principal.User, error)

	CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error
	IssueToken(ctx context.Context, user *principal.User) (*auth.TokenResponse, error)
	IssuePAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error)
}
