package principal

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"github.com/Masterminds/squirrel"
	"gorm.io/gorm"
)

var _ store.IPrincipalStore = (*PrincipalStore)(nil)

func NewPrincipalStore(db *gorm.DB) store.IPrincipalStore {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	return &PrincipalStore{
		db: db,
		sb: factory.CreateQueryBuilder(),
	}
}

type PrincipalStore struct {
	db *gorm.DB
	sb sqlbuilder.QueryBuilder
}

func (s *PrincipalStore) FindUserByID(ctx context.Context, id int64) (*principal.User, error) {
	// 使用 Builder 生成 SQL 语句
	sql, args, err := s.FindUserByIDBuilder(ctx, id).ToSql()
	if err != nil {
		return nil, err
	}

	// 执行查询
	var user Principal
	err = s.db.WithContext(ctx).Raw(sql, args...).Scan(&user).Error
	if err != nil {
		return nil, err
	}

	return toUser(&user), nil
}

func (s *PrincipalStore) FindUserByIDBuilder(ctx context.Context, id int64) sqlbuilder.SelectBuilder {
	return s.sb.
		Select("*").
		From("users").
		Where(squirrel.Eq{"id": id}).
		Limit(1)
}
