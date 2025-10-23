package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"github.com/Masterminds/squirrel"
	"gorm.io/gorm"
)

var _ store.IPrincipalStore = (*PrincipalStore)(nil)

func NewPrincipalStore(db *gorm.DB) store.IPrincipalStore {
	return &PrincipalStore{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

type PrincipalStore struct {
	db *gorm.DB
	sb squirrel.StatementBuilderType
}

func (s *PrincipalStore) FindUserByID(ctx context.Context, id int64) (*model.User, error) {
	// 使用 Builder 生成 SQL 语句
	sql, args, err := s.FindUserByIDBuilder(ctx, id).ToSql()
	if err != nil {
		return nil, err
	}

	// 执行查询
	var user model.User
	err = s.db.WithContext(ctx).Raw(sql, args...).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PrincipalStore) FindUserByIDBuilder(ctx context.Context, id int64) squirrel.SelectBuilder {
	return s.sb.
		Select("*").
		From("users").
		Where(squirrel.Eq{"id": id}).
		Limit(1)
}
