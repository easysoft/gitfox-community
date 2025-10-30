package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

type MySQLSpace struct {
	db *database.MysqlDB
}

func NewMySQLSpace(db *database.MysqlDB) *MySQLSpace {
	return &MySQLSpace{db: db}
}

func (r *MySQLSpace) Create(ctx context.Context, data space.Space) (space.Space, error) {
	// 使用 gorm 的 Create 方法插入数据
	if err := r.db.WithContext(ctx).Create(&data).Error; err != nil {
		return space.Space{}, database.ProcessSQLErrorf(ctx, err)
	}

	return data, nil
}

func (r *MySQLSpace) Update(ctx context.Context, id uint16, data map[string]interface{}) (space.Space, error) {
	var out space.Space

	// 使用事务确保查询和更新的原子性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询出要更新的记录
		if err := tx.Where("id = ?", id).First(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 提取 map 中的所有字段名,用于 Select
		fields := make([]string, 0, len(data))
		for field := range data {
			fields = append(fields, field)
		}

		// 使用 Select 明确指定要更新的字段,支持零值更新
		// Updates 会直接更新 out 对象（传入指针）
		if err := tx.Model(&out).
			Select(fields).
			Updates(data).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		return nil
	})
	if err != nil {
		return space.Space{}, err
	}

	// out 对象已经包含了更新后的数据
	return out, nil
}
