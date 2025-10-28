package sqlbuilder

import (
	"github.com/Masterminds/squirrel"
)

// UniversalQueryBuilder 通用查询构建器
type UniversalQueryBuilder struct {
	dbType DatabaseType
	base   squirrel.StatementBuilderType
}

// NewUniversalBuilder 创建通用查询构建器
func NewUniversalBuilder(dbType DatabaseType) QueryBuilder {
	var placeholderFormat squirrel.PlaceholderFormat

	switch dbType {
	case MySQL, SQLite:
		placeholderFormat = squirrel.Question
	case Postgres:
		placeholderFormat = squirrel.Dollar
	case Dameng:
		placeholderFormat = squirrel.Question // 达梦使用问号占位符
	default:
		placeholderFormat = squirrel.Question
	}

	return &UniversalQueryBuilder{
		dbType: dbType,
		base:   squirrel.StatementBuilder.PlaceholderFormat(placeholderFormat),
	}
}

// GetPlaceholderFormat 获取占位符格式
func (u *UniversalQueryBuilder) GetPlaceholderFormat() squirrel.PlaceholderFormat {
	switch u.dbType {
	case MySQL, SQLite, Dameng:
		return squirrel.Question
	case Postgres:
		return squirrel.Dollar
	default:
		return squirrel.Question
	}
}

// Select 创建查询构建器
func (u *UniversalQueryBuilder) Select(columns ...string) SelectBuilder {
	return &UniversalSelectBuilder{
		dbType:          u.dbType,
		builder:         u.base.Select(columns...),
		conditionActive: true,
	}
}

// Insert 创建插入构建器
func (u *UniversalQueryBuilder) Insert(table string) InsertBuilder {
	return &UniversalInsertBuilder{
		dbType:          u.dbType,
		builder:         u.base.Insert(table),
		conditionActive: true,
	}
}

// Update 创建更新构建器
func (u *UniversalQueryBuilder) Update(table string) UpdateBuilder {
	return &UniversalUpdateBuilder{
		dbType:          u.dbType,
		builder:         u.base.Update(table),
		conditionActive: true,
	}
}

// Delete 创建删除构建器
func (u *UniversalQueryBuilder) Delete(table string) DeleteBuilder {
	return &UniversalDeleteBuilder{
		dbType:  u.dbType,
		builder: u.base.Delete(table),
	}
}
