package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
	"github.com/google/wire"
)

var PubSubProvider = wire.NewSet(
	sse.WireSet,
	pubsub.WireSet,
)
