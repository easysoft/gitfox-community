package scan

import (
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

func newTestSession(account string) *authModel.Session {
	return &authModel.Session{User: &principal.User{Account: account}}
}
