package provider

import (
	"github.com/google/wire"

	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	reviewSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/codereview"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	triggerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/trigger"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
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
	repoStore repoStore.Interface,
	urlProvider url.Provider,
	scanTriggerSvc scanSvc.IScanTrigger,
	planSvc scanSvc.IScanPlan,
	taskSvc scanSvc.IScanTask,
	codeReviewSvc reviewSvc.Interface,
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
		repoStore,
		urlProvider,
		scanTriggerSvc,
		planSvc,
		taskSvc,
		codeReviewSvc,
		gitReaderFactory,
		pullReqEvFactory,
	)
}
