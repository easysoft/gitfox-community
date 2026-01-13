package locker

import "github.com/google/wire"

var WireSet = wire.NewSet(
	// Locker
	NewLocker,
)
