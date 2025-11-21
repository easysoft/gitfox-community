package authn

import (
	"go/types"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
)

func ProvideAuthenticators(
	config *types.Config,
	principalStore principal.Interface,
	tokenStore token.Interface,
) Authenticator {
	jwt := NewJWTAuthenticator(principalStore, tokenStore)
	_ = Register(AuthKindJWT, jwt)

	basic := NewBasicAuthenticator(principalStore)
	_ = Register(AuthKindBasic, basic)

	return jwt
}
