package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var PrincipalWireSet = wire.NewSet(
	NewPrincipalStore,
	wire.Bind(new(principal.IPrincipalStore), new(*principal.PrincipalStore)),
)

func NewPrincipalStore(db *gorm.DB) *principal.PrincipalStore {
	return principal.NewPrincipalStore(db)
}
