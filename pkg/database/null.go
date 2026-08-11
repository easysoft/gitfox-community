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

// BoolToUint8 将 bool 转换为 uint8（true=1，false=0）。
// GaussDB 中对应 MySQL tinyint 的列使用 int 类型，Go bool 无法直接编码，
// 因此模型层与 store 层相互转换时需要显式转换。
func BoolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
