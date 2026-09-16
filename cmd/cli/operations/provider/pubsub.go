package provider

import (
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
	"github.com/google/wire"
)

var PubSubProvider = wire.NewSet(
	sse.WireSet,
	pubsub.WireSet,
)
