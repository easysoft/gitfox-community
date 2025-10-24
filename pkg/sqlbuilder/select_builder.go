package sqlbuilder

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// UniversalSelectBuilder 通用查询构建器实现
type UniversalSelectBuilder struct {
	dbType  DatabaseType
	builder squirrel.SelectBuilder
}

// From 设置表名
func (u *UniversalSelectBuilder) From(table string) SelectBuilder {
	u.builder = u.builder.From(table)
	return u
}

// Where 添加WHERE条件
func (u *UniversalSelectBuilder) Where(pred interface{}, args ...interface{}) SelectBuilder {
	u.builder = u.builder.Where(pred, args...)
	return u
}

// OrderBy 添加排序
func (u *UniversalSelectBuilder) OrderBy(orderBys ...string) SelectBuilder {
	u.builder = u.builder.OrderBy(orderBys...)
	return u
}

// GroupBy 添加分组
func (u *UniversalSelectBuilder) GroupBy(groupBys ...string) SelectBuilder {
	u.builder = u.builder.GroupBy(groupBys...)
	return u
}

// Having 添加HAVING条件
func (u *UniversalSelectBuilder) Having(pred interface{}, args ...interface{}) SelectBuilder {
	u.builder = u.builder.Having(pred, args...)
	return u
}

// Limit 设置限制数量
func (u *UniversalSelectBuilder) Limit(limit uint64) SelectBuilder {
	u.builder = u.builder.Limit(limit)
	return u
}

// Offset 设置偏移量
func (u *UniversalSelectBuilder) Offset(offset uint64) SelectBuilder {
	u.builder = u.builder.Offset(offset)
	return u
}

// Join 添加JOIN
func (u *UniversalSelectBuilder) Join(join string, rest ...interface{}) SelectBuilder {
	u.builder = u.builder.Join(join, rest...)
	return u
}

// LeftJoin 添加LEFT JOIN
func (u *UniversalSelectBuilder) LeftJoin(join string, rest ...interface{}) SelectBuilder {
	u.builder = u.builder.LeftJoin(join, rest...)
	return u
}

// RightJoin 添加RIGHT JOIN
func (u *UniversalSelectBuilder) RightJoin(join string, rest ...interface{}) SelectBuilder {
	u.builder = u.builder.RightJoin(join, rest...)
	return u
}

// InnerJoin 添加INNER JOIN
func (u *UniversalSelectBuilder) InnerJoin(join string, rest ...interface{}) SelectBuilder {
	u.builder = u.builder.InnerJoin(join, rest...)
	return u
}

// ToSql 生成SQL语句
func (u *UniversalSelectBuilder) ToSql() (string, []interface{}, error) {
	return u.builder.ToSql()
}

// ToSqlWithPagination 生成带分页的SQL语句
func (u *UniversalSelectBuilder) ToSqlWithPagination(page, size int) (string, []interface{}, error) {
	switch u.dbType {
	case Dameng:
		return u.toDamengPaginationSql(page, size)
	case MySQL, SQLite, Postgres:
		// 标准的LIMIT OFFSET语法
		u.builder = u.builder.Limit(uint64(size)).Offset(uint64((page - 1) * size))
		return u.builder.ToSql()
	default:
		// 默认使用标准语法
		u.builder = u.builder.Limit(uint64(size)).Offset(uint64((page - 1) * size))
		return u.builder.ToSql()
	}
}

// toDamengPaginationSql 生成达梦数据库分页SQL
func (u *UniversalSelectBuilder) toDamengPaginationSql(page, size int) (string, []interface{}, error) {
	sql, args, err := u.builder.ToSql()
	if err != nil {
		return "", nil, err
	}

	// 达梦数据库分页语法：使用LIMIT offset, count (与MySQL相同)
	// 这与PostgreSQL的 LIMIT count OFFSET offset 不同
	offset := (page - 1) * size
	
	// 应用达梦特有的SQL语法转换
	damengBuilder := NewDamengBuilder()
	sql = damengBuilder.ConvertMySQLToDamengSQL(sql)
	
	// 添加达梦分页语法
	damengSQL := fmt.Sprintf("%s LIMIT %d, %d", sql, offset, size)
	
	return damengSQL, args, nil
}