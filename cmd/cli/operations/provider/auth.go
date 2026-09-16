package provider

import (
	"github.com/easysoft/gitfox-community/pkg/auth/authn"
	"github.com/easysoft/gitfox-community/pkg/auth/authz"
	"github.com/google/wire"
)

var AuthWireSet = wire.NewSet(
	authn.AuthnWireSet,
	authz.AuthzWireSet,
)
