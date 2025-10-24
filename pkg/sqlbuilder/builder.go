package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	MySQL    DatabaseType = "mysql"
	SQLite   DatabaseType = "sqlite"
	Dameng   DatabaseType = "dameng"
	Postgres DatabaseType = "postgres"
)

// QueryBuilder 通用查询构建器接口
type QueryBuilder interface {
	Select(columns ...string) SelectBuilder
	Insert(table string) InsertBuilder
	Update(table string) UpdateBuilder
	Delete(table string) DeleteBuilder
	GetPlaceholderFormat() squirrel.PlaceholderFormat
}

// SelectBuilder 查询构建器接口
type SelectBuilder interface {
	From(table string) SelectBuilder
	Where(pred interface{}, args ...interface{}) SelectBuilder
	OrderBy(orderBys ...string) SelectBuilder
	GroupBy(groupBys ...string) SelectBuilder
	Having(pred interface{}, args ...interface{}) SelectBuilder
	Limit(limit uint64) SelectBuilder
	Offset(offset uint64) SelectBuilder
	Join(join string, rest ...interface{}) SelectBuilder
	LeftJoin(join string, rest ...interface{}) SelectBuilder
	RightJoin(join string, rest ...interface{}) SelectBuilder
	InnerJoin(join string, rest ...interface{}) SelectBuilder
	ToSql() (string, []interface{}, error)
	ToSqlWithPagination(page, size int) (string, []interface{}, error)
}

// InsertBuilder 插入构建器接口
type InsertBuilder interface {
	Columns(columns ...string) InsertBuilder
	Values(values ...interface{}) InsertBuilder
	SetMap(clauses map[string]interface{}) InsertBuilder
	ToSql() (string, []interface{}, error)
}

// UpdateBuilder 更新构建器接口
type UpdateBuilder interface {
	Set(column string, value interface{}) UpdateBuilder
	SetMap(clauses map[string]interface{}) UpdateBuilder
	Where(pred interface{}, args ...interface{}) UpdateBuilder
	ToSql() (string, []interface{}, error)
}

// DeleteBuilder 删除构建器接口
type DeleteBuilder interface {
	Where(pred interface{}, args ...interface{}) DeleteBuilder
	ToSql() (string, []interface{}, error)
}
