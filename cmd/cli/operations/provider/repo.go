package provider


import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
)

var RepoeWireSet = wire.NewSet(
	repo.NewRepoController,
	wire.Bind(new(repoSvc.repoInterface), new(*repoSvc.RepoSvc)),
	// 提供创建实现类实例的方法（关键修复）
	repoSvc.NewRepoSvc,
	wire.Bind(new(repoStore.RepoInterface), new(*repoStore.RepoStore)),
	spaceStore.NewRepoStore,
)
