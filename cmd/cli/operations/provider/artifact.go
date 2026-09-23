// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package provider

import (
	"github.com/google/wire"

	artifactController "github.com/easysoft/gitfox-community/app/api/controller/artifact"
	ctnCtrl "github.com/easysoft/gitfox-community/app/artifact/controller/container"
	ctnHandler "github.com/easysoft/gitfox-community/app/artifact/handler"
	artifactSvc "github.com/easysoft/gitfox-community/app/service/artifact"
	jobsSvc "github.com/easysoft/gitfox-community/app/service/jobs"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	spaceStore "github.com/easysoft/gitfox-community/app/store/space"
	"github.com/easysoft/gitfox-community/pkg/auth/authz"
	"github.com/easysoft/gitfox-community/pkg/config"
	urlPkg "github.com/easysoft/gitfox-community/pkg/utils/url"
)

var ArtifactWireSet = wire.NewSet(
	artifactController.NewController,
	wire.Bind(new(artifactSvc.Interface), new(*artifactSvc.Service)),
	artifactSvc.NewService,
	wire.Bind(new(artifactStore.Interface), new(*artifactStore.Store)),
	artifactStore.NewStore,
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
func ProvideContainerController(
	artStore artifactStore.Interface,
	svc artifactSvc.Interface,
	urlProvider urlPkg.Provider,
	authorizer authz.Authorizer,
) *ctnCtrl.Controller {
	return ctnCtrl.NewController(artStore, svc.FileStore(), urlProvider, authorizer)
}

// ProvideContainerHandler 创建容器镜像 Handler。
func ProvideContainerHandler(
	ctrl *ctnCtrl.Controller,
	svc artifactSvc.Interface,
	artStore artifactStore.Interface,
	spStore spaceStore.Interface,
) *ctnHandler.ContainerHandler {
	return ctnHandler.NewContainerHandler(ctrl, svc, artStore, spStore)
}
