package provider

import (
	publickeySvc "github.com/easysoft/gitfox-community/app/service/publickey"
	publickeyStore "github.com/easysoft/gitfox-community/app/store/publickey"
	"github.com/google/wire"
)

var PublicKeyWireSet = wire.NewSet(
	publickeyStore.NewPublicKeyStore,
	publickeySvc.NewPublicKeyService,
)
