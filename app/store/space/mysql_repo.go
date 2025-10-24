package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gorm.io/gorm"
)

type MySQLSpace struct {
	db *gorm.DB // GORM数据库实例
}

func NewMySQLSpace(db *gorm.DB) *MySQLSpace {
	return &MySQLSpace{db: db}
}

func (r *MySQLSpace) Create(ctx context.Context, req space.Space) (space.Space, error) {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return space.Space{}, err
	}
	return req, nil
}