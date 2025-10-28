package space

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
)

func TestPublicKeyStore_ListByFingerprintBuilder_SQL(t *testing.T) {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	store := &MySQLSpace{
		db: nil,
		sb: factory.CreateQueryBuilder(),
	}

	ctx := context.Background()
	data := space.Space{}
	builder := store.CreateBuilder(ctx, data)

	sql, args, err := builder.ToSql()
	if err != nil {
		t.Fatalf("生成 SQL 失败: %v", err)
	}

	t.Logf("生成的 SQL: %s", sql)
	t.Logf("参数: %v", args)
}
