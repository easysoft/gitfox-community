package provider

import (
	"github.com/google/wire"

	providerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/provider"
)

var ProviderWireSet = wire.NewSet(
	wire.Bind(new(providerStore.Interface), new(*providerStore.Store)),
	providerStore.NewProviderStore,
)
