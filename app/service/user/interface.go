package user

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

type Interface interface {
	FindByID(ctx context.Context, id int64) (*principal.User, error)
	FindByAccount(ctx context.Context, account string) (*principal.User, error)
	FindByEmail(ctx context.Context, email string) (*principal.User, error)

	CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error

	IssueUserPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error)
	IssueAppPAT(ctx context.Context, user *principal.User, identifier string, lifetime *time.Duration) (*auth.TokenResponse, error)
	IssueAdminPAT() (string, error)
}
