package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"gorm.io/gorm"
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
	var out space.Space

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 插入
		insSQL, insArgs, err := r.CreateBuilder(ctx, data).ToSql()
		if err != nil {
			return err
		}
		if err := tx.Exec(insSQL, insArgs...).Error; err != nil {
			return err
		}

		// 2) 获取自增ID（与插入同一连接，保证 LAST_INSERT_ID 正确）
		var lastID uint64
		if err := tx.Raw("SELECT LAST_INSERT_ID() AS id").Scan(&lastID).Error; err != nil {
			return err
		}

		// 3) 查询新创建的数据
		selSQL, selArgs, err := r.GetBuilder(ctx, uint16(lastID)).ToSql()
		if err != nil {
			return err
		}
		if err := tx.Raw(selSQL, selArgs...).Scan(&out).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return space.Space{}, err
	}
	return out, nil
}

func (r *MySQLSpace) CreateBuilder(ctx context.Context, data space.Space) sqlbuilder.InsertBuilder {
	return r.sb.Insert(data.TableName()).
		Columns("`name`", "`desc`", "`owner`", "`acl`", "`createdBy`", "`editedBy`").
		Values(data.Name, data.Desc, data.Owner, data.Permission, data.CreatedBy, data.EditedBy)
}

// GetBuilder 构建根据主键 ID 查询 Space 的 SQL
func (r *MySQLSpace) GetBuilder(ctx context.Context, id uint16) sqlbuilder.SelectBuilder {
	// 显式列，避免与关键字冲突，并保持字段映射稳定
	return r.sb.Select(
		"`id`",
		"`name`",
		"`owner`",
		"`acl`",
		"`desc`",
		"`createdBy`",
		"`createdDate`",
		"`editedBy`",
		"`editedDate`",
		"`deleted`",
	).
		From(space.Space{}.TableName()).
		Where("`id` = ?", id).
		Limit(1)
}
