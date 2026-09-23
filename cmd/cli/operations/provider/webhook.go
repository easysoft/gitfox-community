// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/webhook"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	webhookSvc "github.com/easysoft/gitfox-community/app/service/webhook"
	pullreqStore "github.com/easysoft/gitfox-community/app/store/pullrequest"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	userStore "github.com/easysoft/gitfox-community/app/store/user"
	webhookStore "github.com/easysoft/gitfox-community/app/store/webhook"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
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
	devopsConfig config.DevopsConfig,
	git git.Interface,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	prReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
	webhookStore webhookStore.Interface,
	repoStore repoStore.Interface,
	repoSvc repoSvc.Interface,
	pullreqStore pullreqStore.Interface,
	userStore userStore.Interface,
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
		devopsConfig,
		git,
		gitReaderFactory,
		prReaderFactory,
		webhookStore,
		repoStore,
		repoSvc,
		pullreqStore,
		userStore,
		urlProvider,
	)
}
