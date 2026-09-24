// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package user

import (
	"context"

	principal "github.com/easysoft/gitfox-community/app/model/user"
	"github.com/easysoft/gitfox-community/app/store/token"
	"github.com/easysoft/gitfox-community/app/store/user"

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
