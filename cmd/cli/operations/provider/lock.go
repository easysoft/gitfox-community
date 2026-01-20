package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/locker"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
)

var LockWireSet = wire.NewSet(
	lock.WireSet,
	locker.WireSet,
)
