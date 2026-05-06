package provider

import (
	"github.com/google/wire"

	artifactController "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	ctnCtrl "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/controller/container"
	ctnHandler "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/handler"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	artifactRbac "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac/artifact"
)

var ArtifactWireSet = wire.NewSet(
	artifactController.NewController,
	wire.Bind(new(artifactSvc.Interface), new(*artifactSvc.Service)),
	artifactSvc.NewService,
	wire.Bind(new(artifactStore.Interface), new(*artifactStore.Store)),
	artifactStore.NewStore,
	artifactRbac.WireSet,
	wire.Bind(new(authorizer.Authorizer), new(*authorizer.ArtifactAuthorizer)),
	authorizer.New,
	ProvideArtifactConfig,
	ProvideArtifactJobScheduler,
	ProvideContainerController,
	ProvideContainerHandler,
)

// ProvideArtifactConfig 从全局配置中提取制品库子配置。
func ProvideArtifactConfig(cfg *config.Config) *config.ArtifactConfig {
	cfg.Artifact.Sanitize()
	return &cfg.Artifact
}

// ProvideArtifactJobScheduler 将 jobs.Scheduler 适配为 artifact.JobScheduler 接口。
// 使用 provider 函数而非 wire.Bind，因为 *jobsSvc.Scheduler 由 JobWireSet 提供，
// 不在当前 ArtifactWireSet 中，wire.Bind 要求具体类型在同一 ProviderSet 内。
func ProvideArtifactJobScheduler(s *jobsSvc.Scheduler) artifactSvc.JobScheduler {
	return s
}

// ProvideContainerController 创建容器镜像控制器。
func ProvideContainerController(artStore artifactStore.Interface, svc artifactSvc.Interface) *ctnCtrl.Controller {
	return ctnCtrl.NewController(artStore, svc.FileStore())
}

// ProvideContainerHandler 创建容器镜像 Handler。
func ProvideContainerHandler(
	ctrl *ctnCtrl.Controller,
	svc artifactSvc.Interface,
	artStore artifactStore.Interface,
	spStore spaceStore.Interface,
	cfg *config.Config,
) *ctnHandler.ContainerHandler {
	return ctnHandler.NewContainerHandler(ctrl, svc, artStore, spStore, nil, cfg)
}
