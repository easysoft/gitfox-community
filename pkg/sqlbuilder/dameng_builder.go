package sqlbuilder

import (
	"fmt"
	"regexp"

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

// SelectWithPagination 创建带分页的查询构建器
func (d *DamengQueryBuilder) SelectWithPagination(columns []string, page, size int) squirrel.SelectBuilder {
	// 处理达梦特有的分页逻辑
	return d.base.Select(columns...)
}

// BuildPagedQuery 构建分页查询 - 达梦使用 LIMIT offset, count 语法
func (d *DamengQueryBuilder) BuildPagedQuery(builder squirrel.SelectBuilder, page, size int) (string, []interface{}, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}

	// 达梦分页语法: LIMIT offset, count (与MySQL相同)
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

// ConvertToPagedQuery 将标准查询转换为达梦分页查询
func (d *DamengQueryBuilder) ConvertToPagedQuery(sql string, args []interface{}, page, size int) (string, []interface{}, error) {
	offset := (page - 1) * size

	// 检查是否已经包含LIMIT子句
	if containsLimit(sql) {
		// 如果已有LIMIT，替换为达梦语法
		damengSQL := replaceLimitWithDameng(sql, offset, size)
		return damengSQL, args, nil
	}

	// 添加达梦分页语法: LIMIT offset, count
	damengSQL := fmt.Sprintf("%s LIMIT %d, %d", sql, offset, size)
	return damengSQL, args, nil
}

// ConvertMySQLToDamengSQL 将MySQL特有的SQL语法转换为达梦语法
func (d *DamengQueryBuilder) ConvertMySQLToDamengSQL(sql string) string {
	// 1. 字符串函数转换
	sql = d.convertStringFunctions(sql)

	// 2. 日期函数转换
	sql = d.convertDateFunctions(sql)

	// 3. 其他函数转换
	sql = d.convertOtherFunctions(sql)

	return sql
}

// convertStringFunctions 转换字符串函数
func (d *DamengQueryBuilder) convertStringFunctions(sql string) string {
	// MySQL的CONCAT函数在达梦中也支持，但达梦还支持||操作符
	// MySQL: CONCAT(str1, str2, str3)
	// 达梦: CONCAT(str1, str2, str3) 或 str1 || str2 || str3

	// MySQL的LENGTH函数对应达梦的CHAR_LENGTH
	// MySQL: LENGTH(str) - 返回字节长度
	// 达梦: CHAR_LENGTH(str) - 返回字符长度, BIT_LENGTH(str) - 返回位长度
	sql = regexp.MustCompile(`\bLENGTH\s*\(`).ReplaceAllString(sql, "CHAR_LENGTH(")

	// MySQL的SUBSTRING函数在达梦中也支持，但达梦还支持SUBSTR
	// MySQL: SUBSTRING(str, pos, len)
	// 达梦: SUBSTRING(str, pos, len) 或 SUBSTR(str, pos, len)

	// MySQL的LOCATE函数对应达梦的INSTR
	// MySQL: LOCATE(substr, str)
	// 达梦: INSTR(str, substr) - 注意参数顺序相反
	locateRegex := regexp.MustCompile(`\bLOCATE\s*\(\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
	sql = locateRegex.ReplaceAllString(sql, "INSTR($2, $1)")

	return sql
}

// convertDateFunctions 转换日期函数
func (d *DamengQueryBuilder) convertDateFunctions(sql string) string {
	// MySQL的NOW()函数在达梦中也支持
	// MySQL: NOW()
	// 达梦: NOW() 或 SYSDATE

	// MySQL的DATE_FORMAT函数对应达梦的TO_CHAR
	// MySQL: DATE_FORMAT(date, format)
	// 达梦: TO_CHAR(date, format)
	sql = regexp.MustCompile(`\bDATE_FORMAT\s*\(`).ReplaceAllString(sql, "TO_CHAR(")

	// MySQL的STR_TO_DATE函数对应达梦的TO_DATE
	// MySQL: STR_TO_DATE(str, format)
	// 达梦: TO_DATE(str, format)
	sql = regexp.MustCompile(`\bSTR_TO_DATE\s*\(`).ReplaceAllString(sql, "TO_DATE(")

	// MySQL的YEAR(), MONTH(), DAY()函数在达梦中也支持
	// 但达梦还支持EXTRACT函数
	// MySQL: YEAR(date)
	// 达梦: YEAR(date) 或 EXTRACT(YEAR FROM date)

	return sql
}

// convertOtherFunctions 转换其他函数
func (d *DamengQueryBuilder) convertOtherFunctions(sql string) string {
	// MySQL的IF函数对应达梦的CASE WHEN
	// MySQL: IF(condition, value_if_true, value_if_false)
	// 达梦: CASE WHEN condition THEN value_if_true ELSE value_if_false END
	ifRegex := regexp.MustCompile(`\bIF\s*\(\s*([^,]+)\s*,\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
	sql = ifRegex.ReplaceAllString(sql, "CASE WHEN $1 THEN $2 ELSE $3 END")

	// MySQL的IFNULL函数对应达梦的NVL
	// MySQL: IFNULL(expr1, expr2)
	// 达梦: NVL(expr1, expr2)
	sql = regexp.MustCompile(`\bIFNULL\s*\(`).ReplaceAllString(sql, "NVL(")

	// MySQL的GROUP_CONCAT函数对应达梦的LISTAGG
	// MySQL: GROUP_CONCAT(expr)
	// 达梦: LISTAGG(expr, ',') WITHIN GROUP (ORDER BY expr)
	groupConcatRegex := regexp.MustCompile(`\bGROUP_CONCAT\s*\(\s*([^)]+)\s*\)`)
	sql = groupConcatRegex.ReplaceAllString(sql, "LISTAGG($1, ',') WITHIN GROUP (ORDER BY $1)")

	// MySQL的LIMIT语法转换
	// MySQL: LIMIT count OFFSET offset
	// 达梦: LIMIT offset, count
	limitOffsetRegex := regexp.MustCompile(`\bLIMIT\s+(\d+)\s+OFFSET\s+(\d+)`)
	sql = limitOffsetRegex.ReplaceAllString(sql, "LIMIT $2, $1")

	return sql
}

// containsLimit 检查SQL是否包含LIMIT子句
func containsLimit(sql string) bool {
	// 使用正则表达式检查LIMIT子句
	matched, _ := regexp.MatchString(`(?i)\bLIMIT\b`, sql)
	return matched
}

// replaceLimitWithDameng 替换LIMIT语法为达梦格式
func replaceLimitWithDameng(sql string, offset, size int) string {
	// 替换 LIMIT count OFFSET offset 为 LIMIT offset, count
	limitOffsetRegex := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)\s+OFFSET\s+(\d+)`)
	if limitOffsetRegex.MatchString(sql) {
		return limitOffsetRegex.ReplaceAllString(sql, fmt.Sprintf("LIMIT %d, %d", offset, size))
	}

	// 替换 LIMIT count 为 LIMIT 0, count (如果没有指定offset)
	limitRegex := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)`)
	if limitRegex.MatchString(sql) {
		return limitRegex.ReplaceAllString(sql, fmt.Sprintf("LIMIT %d, %d", offset, size))
	}

	return sql
}

// GetDamengSpecificFeatures 获取达梦数据库特有功能
func (d *DamengQueryBuilder) GetDamengSpecificFeatures() map[string]string {
	return map[string]string{
		"pagination":      "LIMIT offset, count",
		"string_concat":   "CONCAT() or ||",
		"string_length":   "CHAR_LENGTH() for character count, BIT_LENGTH() for bit count",
		"string_position": "INSTR(str, substr)",
		"date_format":     "TO_CHAR(date, format)",
		"date_parse":      "TO_DATE(str, format)",
		"conditional":     "CASE WHEN ... THEN ... ELSE ... END",
		"null_handling":   "NVL(expr1, expr2)",
		"group_concat":    "LISTAGG(expr, separator) WITHIN GROUP (ORDER BY expr)",
		"row_number":      "ROW_NUMBER() OVER (ORDER BY ...)",
		"top_n":           "TOP n or LIMIT n",
		"fetch_next":      "OFFSET n ROWS FETCH NEXT m ROWS ONLY",
		"rownum":          "ROWNUM (Oracle-style)",
	}
}
