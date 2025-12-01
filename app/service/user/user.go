package user

import (
	"context"

	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"

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
