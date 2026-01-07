package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
)

var RepoWireSet = wire.NewSet(
	repo.NewRepoController,
	wire.Bind(new(repoSvc.Interface), new(*repoSvc.Service)),
	repoSvc.NewRepoSvc,
	wire.Bind(new(repoStore.Interface), new(*repoStore.Store)),
	repoStore.NewRepoStore,
	wire.Bind(new(pullrequest.Interface), new(*pullrequest.Store)),
	pullrequest.NewPullRequestStore,
)
