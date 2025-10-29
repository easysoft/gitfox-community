package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalUpdateBuilder 通用更新构建器实现
type UniversalUpdateBuilder struct {
	dbType          DatabaseType
	builder         squirrel.UpdateBuilder
	conditionActive bool
}

// BeginIF 开始条件判断
func (u *UniversalUpdateBuilder) BeginIF(condition bool) UpdateBuilder {
	u.conditionActive = condition
	return u
}

// Fi 结束条件判断
func (u *UniversalUpdateBuilder) Fi() UpdateBuilder {
	u.conditionActive = true
	return u
}

// Set 设置列值
func (u *UniversalUpdateBuilder) Set(column string, value interface{}) UpdateBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Set(column, value)
	return u
}

// SetMap 使用map设置列值
func (u *UniversalUpdateBuilder) SetMap(clauses map[string]interface{}) UpdateBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.SetMap(clauses)
	return u
}

// Where 添加WHERE条件
func (u *UniversalUpdateBuilder) Where(pred interface{}, args ...interface{}) UpdateBuilder {
	if !u.conditionActive {
		return u
	}
	u.builder = u.builder.Where(pred, args...)
	return u
}

func (u *UniversalUpdateBuilder) EditedDate() UpdateBuilder {
	u.builder = u.builder.SetMap(map[string]interface{}{
		FieldEditedDate: squirrel.Expr("NOW()"),
	})
	return u
}

func (u *UniversalUpdateBuilder) EditedBy(editedBy string) UpdateBuilder {
	u.builder = u.builder.SetMap(map[string]interface{}{
		FieldEditedBy: editedBy,
	})
	return u
}

func (u *UniversalUpdateBuilder) Delete(table string) UpdateBuilder {
	u.builder = u.builder.SetMap(map[string]interface{}{
		FieldDeleted: ValueDeleted,
	})
	return u
}

// ToSql 生成SQL语句
func (u *UniversalUpdateBuilder) ToSql() (string, []interface{}, error) {
	return u.builder.ToSql()
}
