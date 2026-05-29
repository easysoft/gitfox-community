package provider

import (
	"github.com/google/wire"

	notifyStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/notify"
)

var NotifyWireSet = wire.NewSet(
	wire.Bind(new(notifyStore.Interface), new(*notifyStore.Store)),
	notifyStore.NewNotifyStore,
)
