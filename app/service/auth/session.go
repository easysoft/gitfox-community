package auth

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal principal.Principal

	// Metadata contains auth related information (access grants, tokenId, sshKeyId, ...)
	// Metadata Metadata

	// SudoUser is the external authenticated principal
	SudoUser *principal.Principal

	// User is the end principal. default to Principal, override if SudoUser exist
	User *principal.Principal
}
