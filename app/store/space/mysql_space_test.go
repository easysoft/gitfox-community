package space

import (
	"context"
	"testing"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/random"
	"gorm.io/gorm"
)

func TestMysqlSpaceStore_Create_SQL(t *testing.T) {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	store := &MySQLSpace{
		db: nil,
		sb: factory.CreateQueryBuilder(),
	}

	ctx := context.Background()
	data := space.Space{
		Name:       "test-space",
		Desc:       "",
		Owner:      "mzy",
		Permission: "private",
		CreatedBy:  "mzy",
		EditedBy:   "mzy",
	}
	builder := store.CreateBuilder(ctx, data)

	sql, args, err := builder.ToSql()
	if err != nil {
		t.Fatalf("生成 SQL 失败: %v", err)
	}

	t.Logf("生成的 SQL: %s", sql)
	t.Logf("参数: %v", args)
}

func TestMysqlSpaceStore_Create(t *testing.T) {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	db := testDB(t)
	store := &MySQLSpace{
		db: db,
		sb: factory.CreateQueryBuilder(),
	}

	ctx := context.Background()
	data := space.Space{
		Name:       "test-space-" + random.RandString(5),
		Desc:       "",
		Owner:      "ldg",
		Permission: "private",
		CreatedBy:  "ldg",
		EditedBy:   "ldg",
	}
	builder := store.CreateBuilder(ctx, data)

	sql, args, err := builder.ToSql()
	if err != nil {
		t.Fatalf("生成 SQL 失败: %v", err)
	}

	var insertedID uint64
	var lastSpace space.Space
	err = store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 执行插入 - 使用 Exec() 而不是 Error
		if err := tx.Exec(sql, args...).Error; err != nil {
			return err
		}

		// 获取自增 ID
		err = tx.Raw(store.sb.Insert("ops_spaces").GetLastInsertIDSql()).Scan(&insertedID).Error
		if err != nil {
			return err
		}

		sql, args, err = store.sb.Select("name").From("ops_space").LastInsert().ToSql()
		if err != nil {
			return err
		}
		t.Logf("查询 SQL: %s", sql)
		return tx.Raw(sql, args...).Scan(&lastSpace).Error
	})

	if err != nil {
		t.Fatalf("插入失败: %v", err)
	}

	t.Logf("插入成功，新增ID: %d", insertedID)
	t.Logf("插入成功，新增空间: %v", lastSpace)

	t.Logf("生成的 SQL: %s", sql)
	t.Logf("参数: %v", args)
}

func testDB(t *testing.T) *database.MysqlDB {
	t.Helper()
	db, _, err := database.NewMysqlDB(config.DatabaseConfig{
		Driver: "mysql",
		// Datasource:      "root:0831will.DB@tcp(10.0.4.11:3306)/gitfox-community?parseTime=true&loc=Local",
		PrintSQL:        true,
		MaxIdleConn:     10,
		MaxOpenConn:     100,
		ConnMaxLifeTime: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	return db
}
