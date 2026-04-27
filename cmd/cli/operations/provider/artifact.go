//go:build ignore
// +build ignore

package provider

import (
	"github.com/google/wire"

	artifactController "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
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
)

// ProvideArtifactConfig 从全局配置中提取制品库子配置。
func ProvideArtifactConfig(cfg *config.Config) *config.ArtifactConfig {
	cfg.Artifact.Sanitize()
	return &cfg.Artifact
}
