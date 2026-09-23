// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package database

import "unicode"

// SQLRewriter 数据库方言 SQL 重写接口。
// 不同数据库对标识符大小写、语法有差异，通过该接口在 SQL 执行前重写以适配方言。
type SQLRewriter interface {
	// RewriteSQL 返回重写后的 SQL。
	RewriteSQL(sql string) string
}

// GaussDBSQLRewriter 是 SQLRewriter 的 GaussDB 实现。
// 该重写器给未加引号的驼峰标识符或保留字列名加双引号。
type GaussDBSQLRewriter struct{}

// gaussdbReservedColumnWords 是必须加引号才能作为列名使用的 GaussDB 保留字。
var gaussdbReservedColumnWords = map[string]struct{}{
	"online": {},
	"desc":   {},
}

// RewriteSQL 给 SQL 中未加引号的驼峰标识符或 GaussDB 保留字列名加双引号。
// 跳过单引号字符串字面量与已加双引号的标识符，避免误改字符串值或重复加引号。
// 特殊处理 ORDER BY 子句：其中的 desc/asc 是排序方向关键字，不加引号。
func (GaussDBSQLRewriter) RewriteSQL(sql string) string {
	var out []byte
	i, n := 0, len(sql)
	inOrderBy := false // 是否处于 ORDER BY 子句
	lastWord := ""     // 上一个标识符（大写），用于识别 "ORDER BY"
	for i < n {
		c := sql[i]
		switch {
		case c == '\'':
			// 单引号字符串字面量：原样复制到下一个未转义的单引号
			start := i
			i++
			for i < n {
				if sql[i] == '\'' {
					if i+1 < n && sql[i+1] == '\'' { // 转义 '' 表示一个单引号
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			out = append(out, sql[start:i]...)
		case c == '"':
			// 已加双引号的标识符：原样复制，不重复加引号
			start := i
			i++
			for i < n {
				if sql[i] == '"' {
					if i+1 < n && sql[i+1] == '"' { // 转义 "" 表示一个双引号
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			out = append(out, sql[start:i]...)
		case isIdentifierStart(sql, i):
			// 标识符（字母/下划线开头），整体读取
			start := i
			for i < n && isIdentifierChar(sql[i]) {
				i++
			}
			word := sql[start:i]
			upper := gaussDBUpper(word)
			// 更新 ORDER BY 上下文
			switch upper {
			case "ORDER":
				inOrderBy = false // 等待随后的 BY
			case "BY":
				if lastWord == "ORDER" {
					inOrderBy = true
				}
			default:
				if inOrderBy && isSQLClauseKeyword(upper) {
					inOrderBy = false // 遇到其他子句关键字，退出 ORDER BY 上下文
				}
			}
			// desc/asc 作为 ORDER BY 排序方向时不加引号
			isOrderByDirection := inOrderBy && (upper == "DESC" || upper == "ASC")
			if (isCamelCase(word) || isGaussDBReservedColumnWord(word)) && !isOrderByDirection {
				out = append(out, '"')
				out = append(out, word...)
				out = append(out, '"')
			} else {
				out = append(out, word...)
			}
			lastWord = upper
		default:
			out = append(out, c)
			i++
		}
	}
	return string(out)
}

// gaussDBUpper 返回单词的大写形式（用于识别 SQL 关键字）。
func gaussDBUpper(word string) string {
	var buf []byte
	for i := 0; i < len(word); i++ {
		c := word[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		buf = append(buf, c)
	}
	return string(buf)
}

// isSQLClauseKeyword 判断单词是否为会结束 ORDER BY 子句的 SQL 子句关键字。
func isSQLClauseKeyword(upper string) bool {
	switch upper {
	case "SELECT", "FROM", "WHERE", "AND", "OR", "GROUP", "HAVING",
		"LIMIT", "OFFSET", "UPDATE", "SET", "INSERT", "INTO", "VALUES",
		"JOIN", "LEFT", "INNER", "RIGHT", "ON", "DELETE", "UNION":
		return true
	}
	return false
}

// isGaussDBReservedColumnWord 判断单词是否为需要加引号的 GaussDB 保留字列名。
func isGaussDBReservedColumnWord(word string) bool {
	_, ok := gaussdbReservedColumnWords[word]
	return ok
}

// isIdentifierStart 判断 byte 是否为标识符起始字符（字母或下划线）。
func isIdentifierStart(s string, i int) bool {
	c := s[i]
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isIdentifierChar 判断 byte 是否为标识符组成部分。
func isIdentifierChar(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '$'
}

// isCamelCase 判断单词是否为驼峰命名（同时含大写与小写字母）。
// 全大写（如 SELECT、COUNT）或全小写（如 deleted、name）不视为驼峰，避免误改。
func isCamelCase(word string) bool {
	if len(word) == 0 {
		return false
	}
	var hasUpper, hasLower bool
	for _, r := range word {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}
	return hasUpper && hasLower
}
