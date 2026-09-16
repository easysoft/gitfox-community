package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/service/locker"
	"github.com/easysoft/gitfox-community/pkg/lock"
)

var LockWireSet = wire.NewSet(
	lock.WireSet,
	locker.WireSet,
)
