// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/cmd/cli/operations/provider"
)

// NewApp assembles the community edition application dependency graph.
// Overlay editions provide their own injector and reuse the exported wire sets
// from the provider package together with their own bindings.
func NewApp(mode string) (*App, func(), error) {
	panic(wire.Build(
		provider.ConfigWireSet,
		provider.LogWireSet,
		provider.DBWireSet,
		provider.PrincipalWireSet,
		provider.GitWireSet,
		provider.CommunityZtflowWireSet,
		provider.SpaceWireSet,
		provider.RepoWireSet,
		provider.CommunityPushRequestCleanerWireSet,
		provider.ProviderWireSet,
		provider.BranchWireSet,
		provider.CommunityArchiveWireSet,
		provider.TagWireSet,
		provider.SvnWireSet,
		provider.PipelineWireSet,
		provider.ExecutionWireSet,
		provider.RunnerWireSet,
		provider.DockerRunnerWireSet,
		provider.ScanRuleSetWireSet,
		provider.ScanRuleWireSet,
		provider.CommunityAIReviewWireSet,
		provider.FitWireSet,
		provider.CommitWireSet,
		provider.WebhookWireSet,
		provider.CronWireSet,
		provider.StatisticsWireSet,
		provider.ControllersWireSet,
		provider.URLWireSet,
		provider.I18nWireSet,
		provider.SSHWireSet,
		provider.LockWireSet,
		provider.EventWireSet,
		provider.GitEventWireSet,
		provider.PullReqEventWireSet,
		provider.PipelineEventWireSet,
		provider.TriggerWireSet,
		provider.PubSubProvider,
		provider.JobWireSet,
		provider.AuthWireSet,
		provider.StorageWireSet,
		provider.ArtifactWireSet,
		wire.Struct(new(App), "*")),
	)
}
