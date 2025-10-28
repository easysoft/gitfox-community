package sqlbuilder

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// DamengQueryBuilder 达梦数据库专用查询构建器
type DamengQueryBuilder struct {
	base squirrel.StatementBuilderType
}

// NewDamengBuilder 创建达梦查询构建器
func NewDamengBuilder() *DamengQueryBuilder {
	return &DamengQueryBuilder{
		base: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}
}

// BuildPagedQuery 构建分页查询 - 达梦使用 LIMIT offset, count 语法
func (d *DamengQueryBuilder) BuildPagedQuery(builder squirrel.SelectBuilder, page, size int) (string, []interface{}, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}

	offset := (page - 1) * size
	damengSQL := fmt.Sprintf("%s LIMIT %d, %d", sql, offset, size)

	return damengSQL, args, nil
}

// BuildRowNumberPagination 使用ROW_NUMBER()方式分页（适用于复杂查询）
func (d *DamengQueryBuilder) BuildRowNumberPagination(builder squirrel.SelectBuilder, page, size int) (string, []interface{}, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}

	offset := (page - 1) * size
	endRow := page * size

	// 达梦支持ROW_NUMBER()窗口函数进行分页
	damengSQL := fmt.Sprintf(`
		SELECT * FROM (
			SELECT ROW_NUMBER() OVER (ORDER BY (SELECT 1)) as rn, t.* FROM (%s) t
		) WHERE rn > %d AND rn <= %d
	`, sql, offset, endRow)

	return damengSQL, args, nil
}
