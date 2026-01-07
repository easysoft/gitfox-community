package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	pullrequestSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pullrequest"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	pullrequestStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
)

var RepoWireSet = wire.NewSet(
	repo.NewRepoController,
	wire.Bind(new(repoSvc.Interface), new(*repoSvc.Service)),
	repoSvc.NewRepoSvc,
	wire.Bind(new(pullrequestSvc.Interface), new(*pullrequestSvc.Service)),
	pullrequestSvc.NewPullRequestSvc,
	wire.Bind(new(repoStore.Interface), new(*repoStore.Store)),
	repoStore.NewRepoStore,
	wire.Bind(new(pullrequestStore.Interface), new(*pullrequestStore.Store)),
	pullrequestStore.NewPullRequestStore,
)
