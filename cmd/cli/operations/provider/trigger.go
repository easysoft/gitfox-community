package provider

import (
	"github.com/google/wire"

	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	triggerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/trigger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

// TriggerWireSet provides a wire set for trigger service.
var TriggerWireSet = wire.NewSet(
	ProvideTriggerService,
)

// ProvideTriggerService creates a new trigger service.
func ProvideTriggerService(
	eventCtx *EventContext,
	pipelineSvc pipelineSvc.Interface,
	triggerer triggerer.Triggerer,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
) (*triggerSvc.Service, error) {
	cfg := triggerSvc.Config{
		EventReaderName: "trigger",
		Concurrency:     1,
		MaxRetries:      3,
	}
	return triggerSvc.New(
		eventCtx.Ctx,
		cfg,
		pipelineSvc,
		triggerer,
		gitReaderFactory,
		pullReqEvFactory,
	)
}
