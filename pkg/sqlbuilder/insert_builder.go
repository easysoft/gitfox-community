package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalInsertBuilder 通用插入构建器实现
type UniversalInsertBuilder struct {
	dbType          DatabaseType
	builder         squirrel.InsertBuilder
	conditionActive bool
}

// BeginIF 开始条件判断
func (u *UniversalInsertBuilder) BeginIF(condition bool) InsertBuilder {
	u.conditionActive = condition
	return u
}

// Fi 结束条件判断
func (u *UniversalInsertBuilder) Fi() InsertBuilder {
	u.conditionActive = true
	return u
}

// Columns 设置列名
func (u *UniversalInsertBuilder) Columns(columns ...string) InsertBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Columns(columns...)
	return u
}

// Values 设置值
func (u *UniversalInsertBuilder) Values(values ...interface{}) InsertBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Values(values...)
	return u
}

// SetMap 使用map设置列值
func (u *UniversalInsertBuilder) SetMap(clauses map[string]interface{}) InsertBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.SetMap(clauses)
	return u
}

func (u *UniversalInsertBuilder) CreatedDate() InsertBuilder {
	u.builder = u.builder.SetMap(map[string]interface{}{
		FieldCreatedDate: squirrel.Expr("NOW()"),
		FieldEditedDate:  squirrel.Expr("NOW()"),
	})
	return u
}

func (u *UniversalInsertBuilder) CreatedBy(createdBy string) InsertBuilder {
	u.builder = u.builder.SetMap(map[string]interface{}{
		FieldCreatedBy: createdBy,
	})
	return u
}

func (u *UniversalInsertBuilder) GetLastInsertIDSql() string {
	return "SELECT LAST_INSERT_ID()"
}

// ToSql 生成SQL语句
func (u *UniversalInsertBuilder) ToSql() (string, []interface{}, error) {
	// u.CreatedDate()
	return u.builder.ToSql()
}
