package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authz"
	"github.com/google/wire"
)

var AuthWireSet = wire.NewSet(
	authn.AuthnWireSet,
	authz.AuthzWireSet,
)
