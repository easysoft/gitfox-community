package authn

import "fmt"

type authKind string

const (
	AuthKindEasy  = "easy"
	AuthKindJWT   = "jwt"
	AuthKindBasic = "basic"
)

var _authMethods = make(map[authKind]Authenticator)

func Register(kind authKind, authenticator Authenticator) error {
	if _, ok := _authMethods[kind]; ok {
		return fmt.Errorf("conflict authenticator '%s'", kind)
	}
	_authMethods[kind] = authenticator
	return nil
}

func Get(kind authKind) Authenticator {
	return _authMethods[kind]
}
