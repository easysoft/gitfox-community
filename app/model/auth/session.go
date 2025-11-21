package auth

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal principal.User

	// Metadata contains auth related information (access grants, tokenId, sshKeyId, ...)
	// 包含认证相关信息（访问权限、tokenId、sshKeyId等）
	Metadata Metadata

	// User is the end principal. default to Principal, override if SudoUser exist
	User *principal.User
}
