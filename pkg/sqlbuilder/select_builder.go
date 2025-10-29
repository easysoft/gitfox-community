package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalSelectBuilder 通用查询构建器实现
type UniversalSelectBuilder struct {
	dbType          DatabaseType
	conditionActive bool
	builder         squirrel.SelectBuilder
}

// 在BeginIF为true时会激活条件
func (u *UniversalSelectBuilder) BeginIF(condition bool) SelectBuilder {
	u.conditionActive = condition
	return u
}

func (u *UniversalSelectBuilder) Fi() SelectBuilder {
	u.conditionActive = true
	return u
}

// From 设置表名
func (u *UniversalSelectBuilder) From(table string) SelectBuilder {
	u.builder = u.builder.From(table)
	return u
}

// Where 添加WHERE条件
func (u *UniversalSelectBuilder) Where(pred interface{}, args ...interface{}) SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Where(pred, args...)
	return u
}

// LastInsertID 添加WHERE条件
func (u *UniversalSelectBuilder) LastInsert() SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Where("id = LAST_INSERT_ID()")
	return u
}

// OrderBy 添加排序
func (u *UniversalSelectBuilder) OrderBy(orderBys ...string) SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.OrderBy(orderBys...)
	return u
}

// GroupBy 添加分组
func (u *UniversalSelectBuilder) GroupBy(groupBys ...string) SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.GroupBy(groupBys...)
	return u
}

// Having 添加HAVING条件
func (u *UniversalSelectBuilder) Having(pred interface{}, args ...interface{}) SelectBuilder {
	if !u.conditionActive {
		return u
	}
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
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Join(join, rest...)
	return u
}

// LeftJoin 添加LEFT JOIN
func (u *UniversalSelectBuilder) LeftJoin(join string, rest ...interface{}) SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.LeftJoin(join, rest...)
	return u
}

// RightJoin 添加RIGHT JOIN
func (u *UniversalSelectBuilder) RightJoin(join string, rest ...interface{}) SelectBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.RightJoin(join, rest...)
	return u
}

// InnerJoin 添加INNER JOIN
func (u *UniversalSelectBuilder) InnerJoin(join string, rest ...interface{}) SelectBuilder {
	if !u.conditionActive {
		return u
	}
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
		damengBuilder := NewDamengBuilder()
		return damengBuilder.BuildPagedQuery(u.builder, page, size)
	case MySQL, SQLite, Postgres:
		// 标准的LIMIT OFFSET语法
		u.builder = u.builder.Limit(uint64(size)).Offset(uint64((page - 1) * size))
		return u.builder.ToSql()
	default:
		u.builder = u.builder.Limit(uint64(size)).Offset(uint64((page - 1) * size))
		return u.builder.ToSql()
	}
}
