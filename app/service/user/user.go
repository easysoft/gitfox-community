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

func (s *Service) FindByID(ctx context.Context, id int64) (*principal.User, error) {
	return s.userStore.FindByID(ctx, id)
}

func (s *Service) FindByAccount(ctx context.Context, account string) (*principal.User, error) {
	return s.userStore.FindByAccount(ctx, account)
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*principal.User, error) {
	return s.userStore.FindByEmail(ctx, email)
}

func (s *Service) CompareHashAndPassword(ctx context.Context, hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
