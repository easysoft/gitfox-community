package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

type MySQLSpace struct {
	db *database.MysqlDB // GORM数据库实例
}

func NewMySQLSpace(db *database.MysqlDB) *MySQLSpace {
	return &MySQLSpace{db: db}
}

func (r *MySQLSpace) Create(ctx context.Context, data space.Space) (space.Space, error) {
	// 从 context 中获取 logger
	log := logger.GetLogger(ctx)
	log.Debugf("[Store] 开始执行数据库创建操作, space_name=%s", data.Name)

	if err := r.db.WithContext(ctx).Create(&data).Error; err != nil {
		log.Errorf("[Store] 数据库创建失败: %v, space_name=%s", err, data.Name)
		return space.Space{}, err
	}

	log.Debugf("[Store] 数据库创建成功, space_id=%d, space_name=%s", data.ID, data.Name)
	return data, nil
}
