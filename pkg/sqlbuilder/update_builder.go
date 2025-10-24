package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalUpdateBuilder 通用更新构建器实现
type UniversalUpdateBuilder struct {
	dbType  DatabaseType
	builder squirrel.UpdateBuilder
}

// Set 设置列值
func (u *UniversalUpdateBuilder) Set(column string, value interface{}) UpdateBuilder {
	u.builder = u.builder.Set(column, value)
	return u
}

// SetMap 使用map设置列值
func (u *UniversalUpdateBuilder) SetMap(clauses map[string]interface{}) UpdateBuilder {
	u.builder = u.builder.SetMap(clauses)
	return u
}

// Where 添加WHERE条件
func (u *UniversalUpdateBuilder) Where(pred interface{}, args ...interface{}) UpdateBuilder {
	u.builder = u.builder.Where(pred, args...)
	return u
}

// ToSql 生成SQL语句
func (u *UniversalUpdateBuilder) ToSql() (string, []interface{}, error) {
	return u.builder.ToSql()
}