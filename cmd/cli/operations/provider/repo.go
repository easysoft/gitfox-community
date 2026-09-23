// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/repo"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/app/service/encrypt"
	importerSvc "github.com/easysoft/gitfox-community/app/service/importer"
	"github.com/easysoft/gitfox-community/app/service/locker"
	pullrequestSvc "github.com/easysoft/gitfox-community/app/service/pullrequest"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/app/service/sse"
	pullrequestStore "github.com/easysoft/gitfox-community/app/store/pullrequest"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
	svnPkg "github.com/easysoft/gitfox-community/pkg/svn"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
)

var RepoWireSet = wire.NewSet(
	repo.NewRepoController,
	wire.Bind(new(repoSvc.Interface), new(*repoSvc.Service)),
	repoSvc.NewRepoSvc,
	wire.Bind(new(importerSvc.Interface), new(*importerSvc.Repository)),
	importerSvc.NewRepoImporterSvc,
	NotifyWireSet,
	ProvideEncrypter,
	wire.Bind(new(pullrequestSvc.Interface), new(*pullrequestSvc.Service)),
	ProvidePullRequestSvc,
	wire.Bind(new(repoStore.Interface), new(*repoStore.Store)),
	repoStore.NewRepoStore,
	wire.Bind(new(pullrequestStore.Interface), new(*pullrequestStore.Store)),
	pullrequestStore.NewPullRequestStore,
	svnPkg.SvnWireSet,
)

// ProvideEncrypter returns a field encrypter. When no key is configured,
// a no-op (plaintext) encrypter is used.
func ProvideEncrypter(cfg *config.Config) (encrypt.Encrypter, error) {
	return encrypt.Provide(cfg.Encryption.Key, cfg.Encryption.Compat)
}

// ProvidePullRequestSvc creates PullRequestService with the event context for graceful shutdown.
func ProvidePullRequestSvc(
	eventCtx *EventContext,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullReqEvReporter *pullreqevents.Reporter,
	cfg *config.Config,
	gitSvc git.Interface,
	repoSvc repoSvc.Interface,
	repoStore repoStore.Interface,
	locker *locker.Locker,
	pullRequestStore pullrequestStore.Interface,
	sseStreamer sse.Streamer,
	bus pubsub.PubSub,
	urlProvider url.Provider,
	pushReqCleaner pullrequestSvc.PushRequestCleaner,
) (*pullrequestSvc.Service, error) {
	return pullrequestSvc.NewPullRequestSvc(
		eventCtx.Ctx,
		gitReaderFactory,
		pullReqEvFactory,
		pullReqEvReporter,
		cfg,
		gitSvc,
		repoSvc,
		repoStore,
		locker,
		pullRequestStore,
		sseStreamer,
		bus,
		urlProvider,
		pushReqCleaner,
	)
}

// CommunityPushRequestCleanerWireSet binds the community-edition no-op push
// request cleaner. The paid edition replaces this with its own wire set.
var CommunityPushRequestCleanerWireSet = wire.NewSet(
	wire.Bind(new(pullrequestSvc.PushRequestCleaner), new(*pullrequestSvc.NoopPushRequestCleaner)),
	pullrequestSvc.NewNoopPushRequestCleaner,
)
