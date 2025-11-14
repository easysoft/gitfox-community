package user

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	principalStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
)

type Service struct {
	userStore principalStore.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 UserSrv 实例
func NewService(userStore principalStore.Interface) *Service {
	return &Service{
		userStore: userStore,
	}
}

func (s *Service) FindUserByUID(ctx context.Context, uid string) (*principal.User, error) {
	return s.userStore.FindUserByUID(ctx, uid)
}

func (s *Service) FindUserByEmail(ctx context.Context, email string) (*principal.User, error) {
	return s.userStore.FindUserByEmail(ctx, email)
}
