package main

import (
	"fmt"
	"log"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"github.com/Masterminds/squirrel"
)

func main() {
	fmt.Println("=== SQLBuilder 链式编程演示 ===\n")

	// 演示各种常用SQL场景的链式编程
	// demoBasicSelect()
	// demoComplexSelect()
	// demoInsertOperations()
	demoUpdateOperations()
	// demoDeleteOperations()
	// demoPaginationDifferences()
	// demoAdvancedFeatures()
}

// 基础查询演示
func demoBasicSelect() {
	fmt.Println("🔍 1. 基础查询链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 简单查询
	sql, args, _ := builder.
		Select("id", "name", "email").
		From("users").
		ToSql()

	fmt.Printf("简单查询: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 带条件查询
	sql, args, _ = builder.
		Select("*").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		Where(squirrel.Gt{"age": 18}).
		ToSql()

	fmt.Printf("条件查询: %s\n", sql)
	fmt.Printf("参数: %v\n\n", args)
}

// 复杂查询演示
func demoComplexSelect() {
	fmt.Println("🔍 2. 复杂查询链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 多表关联 + 复杂条件
	sql, args, _ := builder.
		Select("u.id", "u.name", "u.email", "p.title", "COUNT(pk.id) as key_count").
		From("users u").
		LeftJoin("principals p ON u.id = p.user_id").
		LeftJoin("public_keys pk ON p.id = pk.principal_id").
		Where(squirrel.And{
			squirrel.Eq{"u.status": "active"},
			squirrel.Or{
				squirrel.Like{"u.name": "%admin%"},
				squirrel.Eq{"u.role": "manager"},
			},
		}).
		GroupBy("u.id", "u.name", "u.email", "p.title").
		Having(squirrel.GtOrEq{"COUNT(pk.id)": 1}).
		OrderBy("u.created_at DESC", "u.name ASC").
		Limit(20).
		ToSql()

	fmt.Printf("复杂查询: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 条件构建演示
	fmt.Println("\n📋 条件构建链式编程:")

	// 使用 beginif 和 fi 函数实现动态条件
	isAdmin := true
	activeOnly := false

	query := builder.
		Select("id", "name", "status").
		From("users").
		BeginIF(isAdmin).Where(squirrel.Eq{"role": "admin"}).Fi().       // isAdmin=true 时激活条件 会执行
		BeginIF(activeOnly).Where(squirrel.Eq{"status": "active"}).Fi(). // activeOnly=false 时不激活条件 不会执行
		Where(squirrel.Eq{"deleted_at": nil})                            // 默认激活

	sql, args, _ = query.
		OrderBy("name").
		ToSql()

	fmt.Printf("动态条件: %s\n", sql)
	fmt.Printf("参数: %v\n\n", args)
}

// 插入操作演示
func demoInsertOperations() {
	fmt.Println("🔍 3. 插入操作链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 单条插入 - 逐个字段
	sql, args, _ := builder.
		Insert("users").
		Columns("name", "email", "status", "created_at").
		Values("张三", "zhangsan@example.com", "active", "NOW()").
		ToSql()

	fmt.Printf("单条插入: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 批量插入
	sql, args, _ = builder.
		Insert("users").
		Columns("name", "email", "status").
		Values("李四", "lisi@example.com", "active").
		Values("王五", "wangwu@example.com", "pending").
		Values("赵六", "zhaoliu@example.com", "active").
		ToSql()

	fmt.Printf("批量插入: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// Map方式插入
	userData := map[string]interface{}{
		"name":       "钱七",
		"email":      "qianqi@example.com",
		"status":     "active",
		"age":        25,
		"created_at": "NOW()",
	}

	sql, args, _ = builder.
		Insert("users").
		SetMap(userData).
		ToSql()

	fmt.Printf("Map插入: %s\n", sql)
	fmt.Printf("参数: %v\n\n", args)
}

// 更新操作演示
func demoUpdateOperations() {
	fmt.Println("🔍 4. 更新操作链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 单字段更新
	sql, args, _ := builder.
		Update("users").
		Set("status", "inactive").
		Where(squirrel.Eq{"id": 123}).
		ToSql()

	fmt.Printf("单字段更新: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 多字段更新
	sql, args, _ = builder.
		Update("users").
		Set("status", "verified").
		Set("verified_at", "NOW()").
		Set("verified_by", "admin").
		Where(squirrel.And{
			squirrel.Eq{"email_verified": true},
			squirrel.Eq{"phone_verified": true},
		}).
		ToSql()

	fmt.Printf("多字段更新: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// Map方式更新
	updateData := map[string]interface{}{
		"last_login":  "NOW()",
		"login_count": squirrel.Expr("login_count + 1"),
		"last_ip":     "192.168.1.100",
		"updated_at":  "NOW()",
	}

	sql, args, _ = builder.
		Update("users").
		SetMap(updateData).
		Where(squirrel.Eq{"id": 456}).
		ToSql()

	fmt.Printf("Map更新: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 测试 EditedDate 和 EditedBy 方法
	fmt.Println("\n📝 EditedDate 和 EditedBy 方法测试:")

	// 使用 EditedDate 方法
	updateBuilder := factory.CreateQueryBuilder()
	sql, args, _ = updateBuilder.
		Update("users").
		Set("name", "新用户名").
		EditedDate().
		Where(squirrel.Eq{"id": 789}).
		ToSql()

	fmt.Printf("EditedDate方法: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 使用 EditedBy 方法
	sql, args, _ = updateBuilder.
		Update("users").
		Set("status", "updated").
		EditedBy("admin_user").
		Where(squirrel.Eq{"id": 101}).
		ToSql()

	fmt.Printf("EditedBy方法: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 同时使用 EditedDate 和 EditedBy 方法
	sql, args, _ = updateBuilder.
		Update("users").
		Set("email", "new@example.com").
		EditedDate().
		EditedBy("system_admin").
		Where(squirrel.Eq{"id": 202}).
		ToSql()

	fmt.Printf("同时使用EditedDate和EditedBy: %s\n", sql)
	fmt.Printf("参数: %v\n\n", args)
}

// 删除操作演示
func demoDeleteOperations() {
	fmt.Println("🔍 5. 删除操作链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 简单删除
	sql, args, _ := builder.
		Delete("users").
		Where(squirrel.Eq{"id": 123}).
		ToSql()

	fmt.Printf("简单删除: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 条件删除
	sql, args, _ = builder.
		Delete("users").
		Where(squirrel.And{
			squirrel.Eq{"status": "inactive"},
			squirrel.Lt{"last_login": "2023-01-01"},
			squirrel.Or{
				squirrel.Eq{"email_verified": false},
				//squirrel.IsNull{"email"},
			},
		}).
		ToSql()

	fmt.Printf("条件删除: %s\n", sql)
	fmt.Printf("参数: %v\n\n", args)
}

// 分页差异演示 - 重点展示达梦与其他数据库的差异
func demoPaginationDifferences() {
	fmt.Println("🔍 6. 分页语法差异演示 (达梦 vs 其他数据库)")
	fmt.Println(strings.Repeat("=", 50))

	page, size := 2, 10

	// MySQL/SQLite 分页
	mysqlFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	mysqlBuilder := mysqlFactory.CreateQueryBuilder()

	mysqlQuery := mysqlBuilder.
		Select("id", "name", "email").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC")

	mysqlSQL, mysqlArgs, _ := mysqlQuery.ToSqlWithPagination(page, size)
	fmt.Printf("MySQL分页: %s\n", mysqlSQL)
	fmt.Printf("参数: %v\n", mysqlArgs)

	// PostgreSQL 分页
	pgFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Postgres)
	pgBuilder := pgFactory.CreateQueryBuilder()

	pgQuery := pgBuilder.
		Select("id", "name", "email").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC")

	pgSQL, pgArgs, _ := pgQuery.ToSqlWithPagination(page, size)
	fmt.Printf("PostgreSQL分页: %s\n", pgSQL)
	fmt.Printf("参数: %v\n", pgArgs)

	// 达梦分页 - 关键差异
	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()

	damengQuery := damengBuilder.
		Select("id", "name", "email").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC")

	damengSQL, damengArgs, err := damengQuery.ToSqlWithPagination(page, size)
	if err != nil {
		log.Printf("达梦分页错误: %v", err)
	}
	fmt.Printf("达梦分页: %s\n", damengSQL)
	fmt.Printf("参数: %v\n", damengArgs)

	fmt.Println("\n🎯 分页语法对比:")
	fmt.Println("   MySQL/SQLite: LIMIT 10 OFFSET 10")
	fmt.Println("   PostgreSQL:   LIMIT 10 OFFSET 10")
	fmt.Println("   达梦:         LIMIT 10, 10")
	fmt.Println("   说明: 达梦使用 MySQL 风格的 LIMIT offset, count 语法\n")
}

// 高级特性演示
func demoAdvancedFeatures() {
	fmt.Println("🔍 7. 高级特性链式编程")
	fmt.Println(strings.Repeat("=", 50))

	factory := sqlbuilder.NewBuilderFactory(sqlbuilder.MySQL)
	builder := factory.CreateQueryBuilder()

	// 子查询
	fmt.Println("📊 子查询演示:")

	subQuery := builder.
		Select("user_id").
		From("user_roles").
		Where(squirrel.Eq{"role": "admin"})

	subSQL, subArgs, _ := subQuery.ToSql()

	mainSQL, mainArgs, _ := builder.
		Select("id", "name", "email").
		From("users").
		Where(fmt.Sprintf("id IN (%s)", subSQL), subArgs...).
		ToSql()

	fmt.Printf("子查询: %s\n", subSQL)
	fmt.Printf("主查询: %s\n", mainSQL)
	fmt.Printf("参数: %v\n", mainArgs)

	// 复杂表达式
	fmt.Println("\n📊 复杂表达式演示:")

	sql, args, _ := builder.
		Select(
			"id",
			"name",
			"email",
			"CASE WHEN age >= 18 THEN '成年' ELSE '未成年' END as age_group",
			"DATE_FORMAT(created_at, '%Y-%m') as created_month",
		).
		From("users").
		Where(squirrel.Expr("DATEDIFF(NOW(), created_at) <= ?", 30)).
		OrderBy("created_at DESC").
		ToSql()

	fmt.Printf("复杂表达式: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	// 聚合查询
	fmt.Println("\n📊 聚合查询演示:")

	sql, args, _ = builder.
		Select(
			"department",
			"COUNT(*) as user_count",
			"AVG(age) as avg_age",
			"MAX(created_at) as latest_join",
		).
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		GroupBy("department").
		Having(squirrel.Gt{"COUNT(*)": 5}).
		OrderBy("user_count DESC").
		ToSql()

	fmt.Printf("聚合查询: %s\n", sql)
	fmt.Printf("参数: %v\n", args)

	fmt.Println("\n🎯 SQLBuilder 链式编程优势:")
	fmt.Println("   ✅ 类型安全的SQL构建")
	fmt.Println("   ✅ 防止SQL注入")
	fmt.Println("   ✅ 动态条件构建")
	fmt.Println("   ✅ 跨数据库兼容")
	fmt.Println("   ✅ 代码可读性高")
	fmt.Println("   ✅ 易于测试和维护")
}
