package database

import "unicode"

// SQLRewriter 数据库方言 SQL 重写接口。
// 不同数据库对标识符大小写、语法有差异，通过该接口在 SQL 执行前重写以适配方言。
type SQLRewriter interface {
	// RewriteSQL 返回重写后的 SQL。
	RewriteSQL(sql string) string
}

// GaussDBSQLRewriter 是 SQLRewriter 的 GaussDB 实现。
// GaussDB 预建表使用驼峰列名（如 isCustom、createdBy），而 PostgreSQL/GaussDB
// 协议对未加引号的标识符会强制折叠为小写，导致 Column "iscustom" does not exist。
// 该重写器给未加引号的驼峰标识符加双引号，保留其大小写。
type GaussDBSQLRewriter struct{}

// RewriteSQL 给 SQL 中未加引号的驼峰标识符加双引号。
// 跳过单引号字符串字面量与已加双引号的标识符，避免误改字符串值或重复加引号。
func (GaussDBSQLRewriter) RewriteSQL(sql string) string {
	var out []byte
	i, n := 0, len(sql)
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
			if isCamelCase(word) {
				out = append(out, '"')
				out = append(out, word...)
				out = append(out, '"')
			} else {
				out = append(out, word...)
			}
		default:
			out = append(out, c)
			i++
		}
	}
	return string(out)
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
