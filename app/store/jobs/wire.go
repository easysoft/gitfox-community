//go:build wireinject
// +build wireinject

package jobs

import (
	"github.com/google/wire"
)

// WireSet provides the wire set for job store.
var WireSet = wire.NewSet(
	NewScheduleJob,
)
