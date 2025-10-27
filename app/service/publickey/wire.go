package publickey

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProviderPublicKey,
)

func ProviderPublicKey(
	publicKeyStore publickey.IPublicKeyStore,
	pCache store.PrincipalInfoCache,
) Service {
	return NewService(publicKeyStore, pCache)
}
