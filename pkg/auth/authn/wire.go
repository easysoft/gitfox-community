package authn

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"github.com/google/wire"
)

var AuthnWireSet = wire.NewSet(
	ProvideAuthenticators,
)

func ProvideAuthenticators(
	cfg *config.Config,
	principalStore principal.Interface,
	tokenStore token.Interface,
) Authenticator {
	jwt := NewJWTAuthenticator(principalStore, tokenStore)
	_ = Register(AuthKindJWT, jwt)

	basic := NewBasicAuthenticator(principalStore)
	_ = Register(AuthKindBasic, basic)

	return jwt
}
