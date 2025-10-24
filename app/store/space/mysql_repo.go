package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type MySQLSpace struct {
	db *database.MysqlDB // GORM数据库实例
}

func NewMySQLSpace(db *database.MysqlDB) *MySQLSpace {
	return &MySQLSpace{db: db}
}

func (r *MySQLSpace) Create(ctx context.Context, req space.Space) (space.Space, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return space.Space{}, err
	}
	return req, nil
}
