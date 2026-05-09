package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/encrypt"
	importerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/importer"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/locker"
	pullrequestSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pullrequest"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
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
	wire.Bind(new(importerSvc.Interface), new(*importerSvc.Repository)),
	importerSvc.NewRepoImporterSvc,
	ProvideEncrypter,
	wire.Bind(new(pullrequestSvc.Interface), new(*pullrequestSvc.Service)),
	ProvidePullRequestSvc,
	wire.Bind(new(repoStore.Interface), new(*repoStore.Store)),
	repoStore.NewRepoStore,
	wire.Bind(new(pullrequestStore.Interface), new(*pullrequestStore.Store)),
	pullrequestStore.NewPullRequestStore,
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
