package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

var StorageWireSet = wire.NewSet(
	storage.WireSet,
)
