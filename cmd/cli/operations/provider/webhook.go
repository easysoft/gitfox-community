package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/webhook"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	webhookSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/webhook"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	pullreqStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	webhookStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

var WebhookWireSet = wire.NewSet(
	webhook.NewController,
	wire.Bind(new(webhookSvc.Interface), new(*webhookSvc.Service)),
	ProvideWebhookService,
	wire.Bind(new(webhookStore.Interface), new(*webhookStore.Store)),
	webhookStore.NewWebhookStore,
)

func ProvideWebhookService(
	eventCtx *EventContext,
	git git.Interface,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	prReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
	webhookStore webhookStore.Interface,
	repoStore repoStore.Interface,
	repoSvc repoSvc.Interface,
	pullreqStore pullreqStore.Interface,
	urlProvider url.Provider,
) (*webhookSvc.Service, error) {
	cfg := webhookSvc.Config{
		EventReaderName:     "webhook",
		Concurrency:         1,
		MaxRetries:          3,
		AllowPrivateNetwork: true,
	}
	return webhookSvc.NewWebhookSvc(
		eventCtx.Ctx,
		cfg,
		git,
		gitReaderFactory,
		prReaderFactory,
		webhookStore,
		repoStore,
		repoSvc,
		pullreqStore,
		urlProvider,
	)
}
