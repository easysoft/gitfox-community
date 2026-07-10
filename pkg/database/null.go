package database

import "database/sql"

//   一、写侧（Go → DB）
//   sql.NullString{String: "hello", Valid: false}  // 写入 NULL，"hello" 被丢弃
//   sql.NullString{String: "", Valid: false}       // 写入 NULL
//   sql.NullString{String: "", Valid: true}        // 写入空串 ''
//   sql.NullString{String: "hello", Valid: true}   // 写入 'hello'

//   二、读侧（DB → Go）
//   // DB 里是 NULL
//   pk.Content.Valid  = false
//   pk.Content.String = ""       // 零值，永远是空串

//   // DB 里是空串 ''
//   pk.Content.Valid  = true
//   pk.Content.String = ""

//	// DB 里是 'ssh-ed25519 AAAA'
//	pk.Content.Valid  = true
//	pk.Content.String = "ssh-ed25519 AAAA"

// KeepNullString 空串→NULL，非空→有效值
func NullIfEmpty(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// NoNullString 无条件保留，仅在你确认业务永远非空时用
func NoNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
