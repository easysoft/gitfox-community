package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
)

type MySQLSpace struct {
	db *database.MysqlDB
	sb sqlbuilder.QueryBuilder
}

func NewMySQLSpace(db *database.MysqlDB) *MySQLSpace {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	return &MySQLSpace{db: db, sb: factory.CreateQueryBuilder()}
}

func (r *MySQLSpace) Create(ctx context.Context, data space.Space) (space.Space, error) {
	sql, args, err := r.CreateBuilder(ctx, data).ToSql()
	if err != nil {
		return space.Space{}, err
	}

	var s space.Space
	err = r.db.WithContext(ctx).Raw(sql, args...).Scan(&s).Error
	if err != nil {
		return space.Space{}, err
	}

	return s, nil
}

func (r *MySQLSpace) CreateBuilder(ctx context.Context, data space.Space) sqlbuilder.InsertBuilder {
	return r.sb.
		Insert("users").
		Columns("name", "email", "status", "created_at").
		Values("张三", "zhangsan@example.com", "active", "NOW()")
}
