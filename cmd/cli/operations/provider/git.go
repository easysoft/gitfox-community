package provider

import (
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/storage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/types"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
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
	wire.Bind(new(git.Interface), new(*git.Service)),
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
) (*git.Service, error) {
	return git.New(
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
	git git.Interface,
	urlProvider url.Provider,
	githookFactory hook.ClientFactory,
) *githook.Controller {
	ctrl := githook.NewController(
		principalStore,
		repoStore,
		git,
		urlProvider,
	)
	// Wire the controller into the factory
	// Wire the controller into the factory
	if fct, ok := githookFactory.(*githook.ControllerClientFactory); ok {
		fct.GithookCtrl = ctrl
		fct.Git = git
	}
	return ctrl
}

func NewGitLocalStore() storage.Store {
	return storage.NewLocalStore()
}
