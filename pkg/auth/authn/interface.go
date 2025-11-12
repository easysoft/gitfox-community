package authn

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
)

// Authenticator is an abstraction of an entity that's responsible for authenticating principals
// that are making calls via HTTP.
// Authenticator是负责对主体进行身份验证的实体的抽象
// 通过HTTP进行调用。
type Authenticator interface {
	/*
	 * Tries to authenticate the acting principal if credentials are available.
	 * 如果凭据可用，则尝试验证代理主体。
	 * Returns:
	 *		(session, nil) 		    - 请求包含身份验证数据且主体已验证
	 *		(nil, ErrNoAuthData)	- 请求不包含任何身份验证数据
	 *		(nil, err)  			- 请求包含身份验证数据但验证失败
	 */
	Authenticate(r *http.Request) (*auth.Session, error)
}

var _ Authenticator = (*BasicAuthenticator)(nil)
