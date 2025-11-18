package publickey

import (
	principalStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	publickeyStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey"
	"github.com/google/wire"
)

// WireSet provides all dependencies for publickey service
var WireSet = wire.NewSet(
	// Public Key Store
	publickeyStore.NewPublicKeyStore,

	// Principal Info Cache
	principalStore.NewPrincipalInfoCache,

	// Public Key Service
	NewPublicKeyService,
)
