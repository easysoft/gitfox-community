package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/locker"
	pullrequestSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pullrequest"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	settingsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/settings"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	branchStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/branch"
	pullrequestStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

var RepoWireSet = wire.NewSet(
	repo.NewRepoController,
	wire.Bind(new(repoSvc.Interface), new(*repoSvc.Service)),
	repoSvc.NewRepoSvc,
	wire.Bind(new(branchStore.Interface), new(*branchStore.Store)),
	branchStore.NewBranchStore,
	wire.Bind(new(settingsSvc.Interface), new(*settingsSvc.Service)),
	NewSettingsService,
	wire.Bind(new(pullrequestSvc.Interface), new(*pullrequestSvc.Service)),
	ProvidePullRequestSvc,
	wire.Bind(new(repoStore.Interface), new(*repoStore.Store)),
	repoStore.NewRepoStore,
	wire.Bind(new(pullrequestStore.Interface), new(*pullrequestStore.Store)),
	pullrequestStore.NewPullRequestStore,
)

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
	)
}
