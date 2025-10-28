package publickey

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
)

func TestPublicKeyStore_ListByFingerprintBuilder_SQL(t *testing.T) {
	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	store := &PublicKeyStore{
		db: nil,
		sb: factory.CreateQueryBuilder(),
	}

	ctx := context.Background()
	fingerprint := "test_fingerprint_123"
	builder := store.ListByFingerprintBuilder(ctx, fingerprint)

	sql, args, err := builder.ToSql()
	if err != nil {
		t.Fatalf("生成 SQL 失败: %v", err)
	}

	t.Logf("生成的 SQL: %s", sql)
	t.Logf("参数: %v", args)
}
