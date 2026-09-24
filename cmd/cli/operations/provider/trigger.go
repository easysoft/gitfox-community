// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/google/wire"

	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	scanSvc "github.com/easysoft/gitfox-community/app/service/scan"
	triggerSvc "github.com/easysoft/gitfox-community/app/service/trigger"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
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
	codeReviewSvc scanSvc.IAIReview,
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
