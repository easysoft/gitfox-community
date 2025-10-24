package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalDeleteBuilder 通用删除构建器实现
type UniversalDeleteBuilder struct {
	dbType  DatabaseType
	builder squirrel.DeleteBuilder
}

// Where 添加WHERE条件
func (u *UniversalDeleteBuilder) Where(pred interface{}, args ...interface{}) DeleteBuilder {
	u.builder = u.builder.Where(pred, args...)
	return u
}

// ToSql 生成SQL语句
func (u *UniversalDeleteBuilder) ToSql() (string, []interface{}, error) {
	return u.builder.ToSql()
}