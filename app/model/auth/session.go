package auth

import principal "github.com/easysoft/gitfox-community/app/model/user"

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal *principal.Principal

	// User is the end principal. default to Principal, override if SudoUser exist
	User *principal.User `json:"user"`
}
