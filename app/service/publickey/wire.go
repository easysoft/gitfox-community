package publickey

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProviderPublicKey,
)

func ProviderPublicKey(
	publicKeyStore store.IPublicKeyStore,
	pCache store.PrincipalInfoCache,
) Service {
	return NewService(publicKeyStore, pCache)
}
