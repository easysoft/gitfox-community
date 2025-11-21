package auth

import principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal *principal.Principal

	// User is the end principal. default to Principal, override if SudoUser exist
	User *principal.User
}
