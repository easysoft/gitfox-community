package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/cron"
)

// CronWireSet provides the cron controller wire set.
var CronWireSet = wire.NewSet(
	cron.NewController,
)
