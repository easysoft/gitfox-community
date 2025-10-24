package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalInsertBuilder 通用插入构建器实现
type UniversalInsertBuilder struct {
	dbType  DatabaseType
	builder squirrel.InsertBuilder
}

// Columns 设置列名
func (u *UniversalInsertBuilder) Columns(columns ...string) InsertBuilder {
	u.builder = u.builder.Columns(columns...)
	return u
}

// Values 设置值
func (u *UniversalInsertBuilder) Values(values ...interface{}) InsertBuilder {
	u.builder = u.builder.Values(values...)
	return u
}

// SetMap 使用map设置列值
func (u *UniversalInsertBuilder) SetMap(clauses map[string]interface{}) InsertBuilder {
	u.builder = u.builder.SetMap(clauses)
	return u
}

// ToSql 生成SQL语句
func (u *UniversalInsertBuilder) ToSql() (string, []interface{}, error) {
	return u.builder.ToSql()
}