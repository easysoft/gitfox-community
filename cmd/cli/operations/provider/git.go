package provider

import (
	"fmt"
	"time"

	"github.com/easysoft/gitfox-community/app/api/controller/git"
	"github.com/easysoft/gitfox-community/app/api/controller/githook"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	gitSvc "github.com/easysoft/gitfox-community/app/service/git"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	settingsSvc "github.com/easysoft/gitfox-community/app/service/settings"
	spaceSvc "github.com/easysoft/gitfox-community/app/service/space"
	"github.com/easysoft/gitfox-community/app/store/pullrequest"
	"github.com/easysoft/gitfox-community/app/store/repo"
	settingsStore "github.com/easysoft/gitfox-community/app/store/settings"
	"github.com/easysoft/gitfox-community/app/store/token"
	principal "github.com/easysoft/gitfox-community/app/store/user"
	"github.com/easysoft/gitfox-community/pkg/cache"
	"github.com/easysoft/gitfox-community/pkg/config"
	gitPkg "github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/git/enum"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/git/storage"
	"github.com/easysoft/gitfox-community/pkg/git/types"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
	"github.com/google/wire"
)

var GitWireSet = wire.NewSet(
	NewGitConfig,
	NewLastCommitCache,
	NewGitLocalStore,
	NewGitClientFactory,
	NewGitAdapter,
	NewGit,
	NewGitHookController,
	NewGitService,
	NewGitController,
	NewSettingsService,
	settingsStore.NewSettingStore,
	wire.Bind(new(gitPkg.Interface), new(*gitPkg.Service)),
	wire.Bind(new(gitSvc.Interface), new(*gitSvc.Service)),
	wire.Bind(new(settingsSvc.Interface), new(*settingsSvc.Service)),
)

func NewLastCommitCache(config types.Config) (cache.Cache[api.CommitEntryKey, *api.Commit], error) {
	cacheDuration := config.LastCommitCache.Duration

	// no need to cache if it's too short
	if cacheDuration < time.Second {
		return api.NoLastCommitCache(), nil
	}

	switch config.LastCommitCache.Mode {
	case enum.LastCommitCacheModeNone:
		return api.NoLastCommitCache(), nil
	case enum.LastCommitCacheModeInMemory:
		return api.NewInMemoryLastCommitCache(cacheDuration), nil
	default:
		return nil, fmt.Errorf("unknown last commit cache mode provided: %q", config.LastCommitCache.Mode)
	}
}

// NewGitConfig 从系统配置创建 Git types.Config
func NewGitConfig(cfg *config.Config) types.Config {
	return types.Config{
		Trace:    cfg.Git.Trace,
		Root:     cfg.Git.Root,
		TmpDir:   cfg.Git.TmpDir,
		HookPath: cfg.Git.HookPath,
	}
}

func NewGitAdapter(
	config types.Config,
	lastCommitCache cache.Cache[api.CommitEntryKey, *api.Commit],
	githookFactory hook.ClientFactory,
) (*api.Git, error) {
	return api.New(
		config,
		lastCommitCache,
		githookFactory,
	)
}

func NewGit(
	config types.Config,
	adapter *api.Git,
	hookClientFactory hook.ClientFactory,
	storage storage.Store,
) (*gitPkg.Service, error) {
	return gitPkg.New(
		config,
		adapter,
		hookClientFactory,
		storage,
	)
}

func NewGitClientFactory() hook.ClientFactory {
	return &githook.ControllerClientFactory{
		GithookCtrl: nil,
		Git:         nil,
	}
}

func NewGitHookController(
	principalStore principal.Interface,
	repoStore repo.Interface,
	pullreqStore pullrequest.Interface,
	tokenStore token.Interface,
	git gitPkg.Interface,
	urlProvider url.Provider,
	githookFactory hook.ClientFactory,
	cfg *config.Config,
	prReporter *pullreqevents.Reporter,
	gitReporter *gitevents.Reporter,
	pushRequest githook.PushRequestFlow,
) *githook.Controller {
	ctrl := githook.NewController(
		principalStore,
		repoStore,
		pullreqStore,
		tokenStore,
		git,
		urlProvider,
		cfg.Devops,
		prReporter,
		gitReporter,
		pushRequest,
	)
	// Wire the controller into the factory
	if fct, ok := githookFactory.(*githook.ControllerClientFactory); ok {
		fct.GithookCtrl = ctrl
		fct.Git = git
	}
	return ctrl
}

// CommunityZtflowWireSet binds the community-edition no-op code push request
// (AGit refs/for) flow. The paid edition replaces this with its own wire set.
var CommunityZtflowWireSet = wire.NewSet(
	wire.Bind(new(githook.PushRequestFlow), new(*githook.NoopPushRequestFlow)),
	githook.NewNoopPushRequestFlow,
)

func NewGitLocalStore() storage.Store {
	return storage.NewLocalStore()
}

func NewGitService(gitPkg gitPkg.Interface, urlProvider url.Provider) *gitSvc.Service {
	return gitSvc.NewGitService(gitPkg, urlProvider).(*gitSvc.Service)
}

func NewSettingsService(settingsStore settingsStore.Interface) *settingsSvc.Service {
	return settingsSvc.NewSettingsService(settingsStore).(*settingsSvc.Service)
}

func NewGitController(gitService gitSvc.Interface, spaceService spaceSvc.Interface, repoService repoSvc.Interface, settingsService settingsSvc.Interface, githookCtrl *githook.Controller, gitPkgInterface gitPkg.Interface) *git.GitController {
	return git.NewGitController(gitService, spaceService, repoService, settingsService, githookCtrl, gitPkgInterface)
}
