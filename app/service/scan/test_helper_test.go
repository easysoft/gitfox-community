package scan

import (
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
)

func newTestSession(account string) *authModel.Session {
	return &authModel.Session{User: &principal.User{Account: account}}
}
