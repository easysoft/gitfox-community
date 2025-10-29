package space

import (
	"context"
	"time"

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
		// 插入
		insSQL, insArgs, err := r.CreateBuilder(ctx, data).ToSql()
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Exec(insSQL, insArgs...).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 查询新创建的数据
		selSQL, selArgs, err := r.GetBuilder(ctx).ToSql()
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Raw(selSQL, selArgs...).Scan(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
	if err != nil {
		return space.Space{}, err
	}
	return out, nil
}

func (r *MySQLSpace) Update(ctx context.Context, id uint16, data map[string]interface{}) (space.Space, error) {
	var out space.Space

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 构建并执行更新
		updateSQL, updateArgs, err := r.UpdateBuilder(ctx, id, data).ToSql()
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Exec(updateSQL, updateArgs...).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 查询更新后的数据
		selSQL, selArgs, err := r.GetBuilderWithID(ctx, id).ToSql()
		if err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		if err := tx.Raw(selSQL, selArgs...).Scan(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
	if err != nil {
		return space.Space{}, err
	}
	return out, nil
}

// GetBuilder 查询新增 Space 的 SQL
func (r *MySQLSpace) GetBuilder(ctx context.Context) sqlbuilder.SelectBuilder {
	return r.sb.Select("*").From(space.Space{}.TableName()).LastInsert()
}

// GetBuilderWithID 查询指定 ID 的 Space 的 SQL
func (r *MySQLSpace) GetBuilderWithID(ctx context.Context, id uint16) sqlbuilder.SelectBuilder {
	return r.sb.Select("*").From(space.Space{}.TableName()).Where(`id = ?`, id)
}

func (r *MySQLSpace) CreateBuilder(ctx context.Context, data space.Space) sqlbuilder.InsertBuilder {
	return r.sb.Insert(data.TableName()).SetMap(map[string]interface{}{
		"`name`":        data.Name,
		"`desc`":        data.Desc,
		"`owner`":       data.Owner,
		"`acl`":         data.Permission,
		"`createdBy`":   data.CreatedBy,
		"`createdDate`": time.Now(),
		"`editedBy`":    data.EditedBy,
		"`editedDate`":  time.Now(),
	})
}

func (r *MySQLSpace) UpdateBuilder(ctx context.Context, id uint16, data map[string]interface{}) sqlbuilder.UpdateBuilder {
	return r.sb.Update(space.Space{}.TableName()).Where("`id` = ?", id).SetMap(data)
}
