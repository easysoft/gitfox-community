package sse

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewSSE,
)
