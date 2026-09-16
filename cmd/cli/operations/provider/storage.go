package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/pkg/storage"
)

var StorageWireSet = wire.NewSet(
	storage.WireSet,
)
