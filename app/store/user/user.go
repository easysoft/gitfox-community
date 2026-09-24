// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package user

import (
	"context"

	principal "github.com/easysoft/gitfox-community/app/model/user"
	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

func NewUserStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

func (s *Store) FindByID(ctx context.Context, id int64) (*principal.User, error) {
	var user User
	user.ID = id
	err := s.db.WithContext(ctx).Select("id, account, email, password, role, realname").First(&user).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return user.toUser(), nil
}

func (s *Store) FindByAccount(ctx context.Context, account string) (*principal.User, error) {
	var user User
	err := s.db.WithContext(ctx).Select("id, account, email, password, role, realname").Where("account = ?", account).First(&user).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return user.toUser(), nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (*principal.User, error) {
	var user User
	err := s.db.WithContext(ctx).Select("id, account, email, password, role, realname").Where("email = ?", email).First(&user).Error

	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return user.toUser(), nil
}
