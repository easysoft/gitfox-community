package main

import (
	"fmt"
	"log"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/sqlbuilder"
	"github.com/Masterminds/squirrel"
)

func main() {
	fmt.Println("=== SQLBuilder Demo: 达梦数据库 vs 标准 SQL ===\n")

	// 演示各种常用场景
	demoBasicSelect()
	demoSelectWithConditions()
	demoSelectWithJoins()
	demoSelectWithPagination()
	demoInsertOperations()
	demoUpdateOperations()
	demoDeleteOperations()
	demoSyntaxDifferences() // 新增：语法差异演示
	demoComplexQueries()
}

// 基础查询演示
func demoBasicSelect() {
	fmt.Println("🔍 1. 基础查询演示")
	fmt.Println(strings.Repeat("=", 50))

	// 达梦数据库构建器
	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()

	// 标准 squirrel 构建器
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 构建相同的查询
	columns := []string{"id", "name", "email", "created_at"}

	// 达梦方式
	damengQuery := damengBuilder.Select(columns...).From("users")
	damengSQL, damengArgs, _ := damengQuery.ToSql()

	// 标准方式
	standardQuery := standardBuilder.Select(columns...).From("users")
	standardSQL, _, _ := standardQuery.ToSql()

	fmt.Printf("📊 达梦 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n\n", damengArgs)
}

// 带条件查询演示
func demoSelectWithConditions() {
	fmt.Println("🔍 2. 带条件查询演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 复杂条件查询
	// 达梦方式
	damengQuery := damengBuilder.
		Select("u.id", "u.name", "u.email", "p.title").
		From("users u").
		Where(squirrel.And{
			squirrel.Eq{"u.status": "active"},
			squirrel.Gt{"u.created_at": "2023-01-01"},
			squirrel.Or{
				squirrel.Like{"u.name": "%admin%"},
				squirrel.Eq{"u.role": "administrator"},
			},
		}).
		OrderBy("u.created_at DESC", "u.name ASC").
		Limit(10)

	damengSQL, damengArgs, _ := damengQuery.ToSql()

	// 标准方式
	standardQuery := standardBuilder.
		Select("u.id", "u.name", "u.email", "p.title").
		From("users u").
		Where(squirrel.And{
			squirrel.Eq{"u.status": "active"},
			squirrel.Gt{"u.created_at": "2023-01-01"},
			squirrel.Or{
				squirrel.Like{"u.name": "%admin%"},
				squirrel.Eq{"u.role": "administrator"},
			},
		}).
		OrderBy("u.created_at DESC", "u.name ASC").
		Limit(10)

	standardSQL, _, _ := standardQuery.ToSql()

	fmt.Printf("📊 达梦 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n\n", damengArgs)
}

// JOIN 查询演示
func demoSelectWithJoins() {
	fmt.Println("🔍 3. JOIN 查询演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 多表关联查询
	// 达梦方式
	damengQuery := damengBuilder.
		Select("u.id", "u.name", "u.email", "p.title", "pk.fingerprint").
		From("users u").
		LeftJoin("principals p ON u.id = p.user_id").
		InnerJoin("public_keys pk ON p.id = pk.principal_id").
		Where(squirrel.And{
			squirrel.Eq{"u.status": "active"},
			squirrel.NotEq{"pk.fingerprint": nil},
		}).
		GroupBy("u.id", "u.name", "u.email").
		Having(squirrel.Gt{"COUNT(pk.id)": 0}).
		OrderBy("u.name")

	damengSQL, damengArgs, _ := damengQuery.ToSql()

	// 标准方式
	standardQuery := standardBuilder.
		Select("u.id", "u.name", "u.email", "p.title", "pk.fingerprint").
		From("users u").
		LeftJoin("principals p ON u.id = p.user_id").
		InnerJoin("public_keys pk ON p.id = pk.principal_id").
		Where(squirrel.And{
			squirrel.Eq{"u.status": "active"},
			squirrel.NotEq{"pk.fingerprint": nil},
		}).
		GroupBy("u.id", "u.name", "u.email").
		Having(squirrel.Gt{"COUNT(pk.id)": 0}).
		OrderBy("u.name")

	standardSQL, _, _ := standardQuery.ToSql()

	fmt.Printf("📊 达梦 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n\n", damengArgs)
}

// 分页查询演示 - 这里是关键差异
func demoSelectWithPagination() {
	fmt.Println("🔍 4. 分页查询演示 (关键差异)")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	page, size := 2, 10

	// 达梦方式 - 自动处理分页语法
	damengQuery := damengBuilder.
		Select("id", "name", "email", "created_at").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC")

	damengSQL, damengArgs, err := damengQuery.ToSqlWithPagination(page, size)
	if err != nil {
		log.Printf("达梦分页查询错误: %v", err)
	}

	// 标准方式 - 手动处理分页
	offset := (page - 1) * size
	standardQuery := standardBuilder.
		Select("id", "name", "email", "created_at").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC").
		Limit(uint64(size)).
		Offset(uint64(offset))

	standardSQL, standardArgs, _ := standardQuery.ToSql()

	fmt.Printf("📊 达梦分页 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准分页 SQL: %s\n", standardSQL)
	fmt.Printf("达梦参数: %v\n", damengArgs)
	fmt.Printf("标准参数: %v\n", standardArgs)
	fmt.Printf("🎯 注意: 达梦使用 'LIMIT offset, count' 语法\n\n")
}

// 插入操作演示
func demoInsertOperations() {
	fmt.Println("🔍 5. 插入操作演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 单条插入
	// 达梦方式
	damengInsert := damengBuilder.
		Insert("users").
		Columns("name", "email", "status", "created_at").
		Values("张三", "zhangsan@example.com", "active", "NOW()")

	damengSQL, damengArgs, _ := damengInsert.ToSql()

	// 标准方式
	standardInsert := standardBuilder.
		Insert("users").
		Columns("name", "email", "status", "created_at").
		Values("张三", "zhangsan@example.com", "active", "NOW()")

	standardSQL, _, _ := standardInsert.ToSql()

	fmt.Printf("📊 达梦插入 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准插入 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n", damengArgs)

	// 批量插入演示
	fmt.Println("\n📦 批量插入演示:")

	userData := map[string]interface{}{
		"name":       "李四",
		"email":      "lisi@example.com",
		"status":     "active",
		"created_at": "NOW()",
	}

	// 达梦方式
	damengBatchInsert := damengBuilder.Insert("users").SetMap(userData)
	damengBatchSQL, damengBatchArgs, _ := damengBatchInsert.ToSql()

	// 标准方式
	standardBatchInsert := standardBuilder.Insert("users").SetMap(userData)
	standardBatchSQL, _, _ := standardBatchInsert.ToSql()

	fmt.Printf("📊 达梦批量插入 SQL: %s\n", damengBatchSQL)
	fmt.Printf("📊 标准批量插入 SQL: %s\n", standardBatchSQL)
	fmt.Printf("参数: %v\n\n", damengBatchArgs)
}

// 更新操作演示
func demoUpdateOperations() {
	fmt.Println("🔍 6. 更新操作演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 条件更新
	// 达梦方式
	damengUpdate := damengBuilder.
		Update("users").
		Set("status", "inactive").
		Set("updated_at", "NOW()").
		Where(squirrel.And{
			squirrel.Lt{"last_login": "2023-01-01"},
			squirrel.Eq{"status": "active"},
		})

	damengSQL, damengArgs, _ := damengUpdate.ToSql()

	// 标准方式
	standardUpdate := standardBuilder.
		Update("users").
		Set("status", "inactive").
		Set("updated_at", "NOW()").
		Where(squirrel.And{
			squirrel.Lt{"last_login": "2023-01-01"},
			squirrel.Eq{"status": "active"},
		})

	standardSQL, _, _ := standardUpdate.ToSql()

	fmt.Printf("📊 达梦更新 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准更新 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n", damengArgs)

	// 批量更新演示
	fmt.Println("\n📦 批量更新演示:")

	updateData := map[string]interface{}{
		"status":      "verified",
		"updated_at":  "NOW()",
		"verified_by": "system",
	}

	// 达梦方式
	damengBatchUpdate := damengBuilder.Update("users").
		SetMap(updateData).
		Where(squirrel.Eq{"email_verified": false})

	damengBatchSQL, damengBatchArgs, _ := damengBatchUpdate.ToSql()

	// 标准方式
	standardBatchUpdate := standardBuilder.Update("users").
		SetMap(updateData).
		Where(squirrel.Eq{"email_verified": false})

	standardBatchSQL, _, _ := standardBatchUpdate.ToSql()

	fmt.Printf("📊 达梦批量更新 SQL: %s\n", damengBatchSQL)
	fmt.Printf("📊 标准批量更新 SQL: %s\n", standardBatchSQL)
	fmt.Printf("参数: %v\n\n", damengBatchArgs)
}

// 删除操作演示
func demoDeleteOperations() {
	fmt.Println("🔍 7. 删除操作演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()
	standardBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// 条件删除
	// 达梦方式
	damengDelete := damengBuilder.
		Delete("users").
		Where(squirrel.And{
			squirrel.Eq{"status": "inactive"},
			squirrel.Lt{"last_login": "2022-01-01"},
		})

	damengSQL, damengArgs, _ := damengDelete.ToSql()

	// 标准方式
	standardDelete := standardBuilder.
		Delete("users").
		Where(squirrel.And{
			squirrel.Eq{"status": "inactive"},
			squirrel.Lt{"last_login": "2022-01-01"},
		})

	standardSQL, _, _ := standardDelete.ToSql()

	fmt.Printf("📊 达梦删除 SQL: %s\n", damengSQL)
	fmt.Printf("📊 标准删除 SQL: %s\n", standardSQL)
	fmt.Printf("参数: %v\n\n", damengArgs)
}

// 语法差异演示
func demoSyntaxDifferences() {
	fmt.Println("🔍 8. 达梦与MySQL语法差异详细对比")
	fmt.Println(strings.Repeat("=", 60))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	// damengBuilder := damengFactory.CreateQueryBuilder()

	// 创建达梦构建器用于语法转换
	damengSpecialBuilder := damengFactory.CreateDamengBuilder()

	// 1. 字符串函数差异演示
	fmt.Println("📝 1. 字符串函数差异对比")
	fmt.Println(strings.Repeat("-", 40))

	mysqlStringSQL := "SELECT LENGTH(name), LOCATE('@', email) FROM users"
	damengStringSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(mysqlStringSQL)

	fmt.Printf("MySQL: %s\n", mysqlStringSQL)
	fmt.Printf("达梦:  %s\n", damengStringSQL)
	fmt.Println("差异说明:")
	fmt.Println("  - LENGTH() → CHAR_LENGTH() (字符长度)")
	fmt.Println("  - LOCATE(substr, str) → INSTR(str, substr) (参数顺序相反)")
	fmt.Println()

	// 2. 日期函数差异演示
	fmt.Println("📅 2. 日期函数差异对比")
	fmt.Println(strings.Repeat("-", 40))

	mysqlDateSQL := "SELECT DATE_FORMAT(created_at, '%Y-%m-%d'), STR_TO_DATE('2023-01-01', '%Y-%m-%d') FROM users"
	damengDateSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(mysqlDateSQL)

	fmt.Printf("MySQL: %s\n", mysqlDateSQL)
	fmt.Printf("达梦:  %s\n", damengDateSQL)
	fmt.Println("差异说明:")
	fmt.Println("  - DATE_FORMAT() → TO_CHAR() (日期格式化)")
	fmt.Println("  - STR_TO_DATE() → TO_DATE() (字符串转日期)")
	fmt.Println()

	// 3. 条件函数差异演示
	fmt.Println("🔀 3. 条件函数差异对比")
	fmt.Println(strings.Repeat("-", 40))

	mysqlCondSQL := "SELECT IF(status = 1, 'Active', 'Inactive'), IFNULL(phone, 'N/A') FROM users"
	damengCondSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(mysqlCondSQL)

	fmt.Printf("MySQL: %s\n", mysqlCondSQL)
	fmt.Printf("达梦:  %s\n", damengCondSQL)
	fmt.Println("差异说明:")
	fmt.Println("  - IF(cond, val1, val2) → CASE WHEN cond THEN val1 ELSE val2 END")
	fmt.Println("  - IFNULL(expr1, expr2) → NVL(expr1, expr2)")
	fmt.Println()

	// 4. 分页语法差异演示
	fmt.Println("📄 4. 分页语法差异对比")
	fmt.Println(strings.Repeat("-", 40))

	mysqlPageSQL := "SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 20"
	damengPageSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(mysqlPageSQL)

	fmt.Printf("MySQL: %s\n", mysqlPageSQL)
	fmt.Printf("达梦:  %s\n", damengPageSQL)
	fmt.Println("差异说明:")
	fmt.Println("  - LIMIT count OFFSET offset → LIMIT offset, count")
	fmt.Println()

	// 5. 聚合函数差异演示
	fmt.Println("📊 5. 聚合函数差异对比")
	fmt.Println(strings.Repeat("-", 40))

	mysqlAggSQL := "SELECT GROUP_CONCAT(name) FROM users GROUP BY department"
	damengAggSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(mysqlAggSQL)

	fmt.Printf("MySQL: %s\n", mysqlAggSQL)
	fmt.Printf("达梦:  %s\n", damengAggSQL)
	fmt.Println("差异说明:")
	fmt.Println("  - GROUP_CONCAT(expr) → LISTAGG(expr, ',') WITHIN GROUP (ORDER BY expr)")
	fmt.Println()

	// 6. 综合示例 - 复杂SQL转换
	fmt.Println("🎯 6. 综合示例 - 复杂SQL语法转换")
	fmt.Println(strings.Repeat("-", 40))

	complexMySQLSQL := `
		SELECT 
			CONCAT(first_name, ' ', last_name) as full_name,
			LENGTH(email) as email_length,
			LOCATE('@', email) as at_position,
			DATE_FORMAT(created_at, '%Y-%m-%d') as formatted_date,
			IF(status = 1, 'Active', 'Inactive') as status_text,
			IFNULL(phone, 'N/A') as phone_display,
			GROUP_CONCAT(role_name) as roles
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		WHERE created_at >= STR_TO_DATE('2023-01-01', '%Y-%m-%d')
		GROUP BY u.id
		ORDER BY created_at DESC
		LIMIT 10 OFFSET 20
	`

	fmt.Println("原始MySQL复杂SQL:")
	fmt.Println(complexMySQLSQL)

	// 转换为达梦语法
	complexDamengSQL := damengSpecialBuilder.ConvertMySQLToDamengSQL(complexMySQLSQL)
	fmt.Println("\n转换后的达梦语法SQL:")
	fmt.Println(complexDamengSQL)

	// 展示主要差异
	fmt.Println("\n🔍 达梦数据库特有功能总览:")
	features := damengSpecialBuilder.GetDamengSpecificFeatures()
	for feature, description := range features {
		fmt.Printf("- %s: %s\n", feature, description)
	}

	// 7. 达梦特有的窗口函数分页演示
	fmt.Println("\n🪟 7. 达梦窗口函数分页演示")
	fmt.Println(strings.Repeat("-", 40))

	baseQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question).
		Select("id", "name", "email", "created_at").
		From("users").
		Where(squirrel.Eq{"status": "active"}).
		OrderBy("created_at DESC")

	specialSQL, specialArgs, err := damengSpecialBuilder.BuildRowNumberPagination(baseQuery, 2, 10)
	if err != nil {
		log.Printf("达梦专用分页错误: %v", err)
	} else {
		fmt.Printf("达梦窗口函数分页 SQL: %s\n", specialSQL)
		fmt.Printf("参数: %v\n", specialArgs)
		fmt.Println("说明: 使用ROW_NUMBER()窗口函数实现分页，适用于复杂查询场景")
	}
	fmt.Println()
}

// 复杂查询演示
func demoComplexQueries() {
	fmt.Println("🔍 9. 复杂查询和高级功能演示")
	fmt.Println(strings.Repeat("=", 50))

	damengFactory := sqlbuilder.NewBuilderFactory(sqlbuilder.Dameng)
	damengBuilder := damengFactory.CreateQueryBuilder()

	// 创建达梦构建器用于语法转换
	// damengSpecialBuilder := damengFactory.CreateDamengBuilder()

	// 子查询示例
	fmt.Println("📊 1. 子查询示例:")

	// 构建子查询
	subQuery := damengBuilder.
		Select("principal_id").
		From("public_keys").
		Where(squirrel.Eq{"usage": "auth"})

	subQuerySQL, subQueryArgs, _ := subQuery.ToSql()

	// 主查询使用子查询
	mainQuery := damengBuilder.
		Select("u.id", "u.name", "u.email").
		From("users u").
		Join("principals p ON u.id = p.user_id").
		Where(fmt.Sprintf("p.id IN (%s)", subQuerySQL), subQueryArgs...)

	mainSQL, mainArgs, _ := mainQuery.ToSql()

	fmt.Printf("子查询 SQL: %s\n", subQuerySQL)
	fmt.Printf("主查询 SQL: %s\n", mainSQL)
	fmt.Printf("参数: %v\n", mainArgs)

	// 窗口函数示例（达梦特色）
	fmt.Println("\n📊 2. 窗口函数示例（达梦特色）:")

	windowQuery := damengBuilder.
		Select(
			"id",
			"name",
			"email",
			"ROW_NUMBER() OVER (PARTITION BY department ORDER BY created_at DESC) as rn",
		).
		From("users").
		Where(squirrel.Eq{"status": "active"})

	windowSQL, windowArgs, _ := windowQuery.ToSql()
	fmt.Printf("窗口函数 SQL: %s\n", windowSQL)
	fmt.Printf("参数: %v\n", windowArgs)

	// CTE (公用表表达式) 示例
	fmt.Println("\n📊 3. CTE (公用表表达式) 示例:")

	cteSQL := `
		WITH active_users AS (
			SELECT id, name, email, department 
			FROM users 
			WHERE status = 'active'
		),
		dept_stats AS (
			SELECT department, COUNT(*) as user_count
			FROM active_users
			GROUP BY department
		)
		SELECT au.name, au.email, ds.user_count
		FROM active_users au
		JOIN dept_stats ds ON au.department = ds.department
		ORDER BY ds.user_count DESC, au.name
	`

	fmt.Printf("CTE SQL: %s\n", cteSQL)
	fmt.Println("说明: 达梦支持WITH子句创建公用表表达式，提高复杂查询的可读性")

	fmt.Println("\n🎯 总结:")
	fmt.Println("✅ 达梦 SQLBuilder 的优势:")
	fmt.Println("   - 自动处理达梦特有的分页语法")
	fmt.Println("   - 支持窗口函数分页")
	fmt.Println("   - 统一的接口，易于切换数据库")
	fmt.Println("   - 保持与标准 squirrel 的兼容性")
	fmt.Println("   - 支持复杂查询和子查询")
	fmt.Println("   - 自动转换MySQL语法为达梦语法")
	fmt.Println("   - 支持多种分页方式（LIMIT和窗口函数）")
	fmt.Println("   - 支持CTE、窗口函数等高级SQL特性")
}
