//go:build wireinject
// +build wireinject

package server

import (
	"github.com/drone/runner-go/poller"
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	triggerSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/trigger"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	tokenStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/ssh"
	"gorm.io/gorm"
)

type App struct {
	SystemConfig    *config.Config
	SystemLog       *provider.SystemLog
	APILog          *provider.APILog
	DB              *gorm.DB
	Controller      *router.Controller
	I18nManager     *provider.I18nManager
	GithookCtrl     *githook.Controller
	SSHServer       *ssh.Server
	UserService     userSvc.Interface
	PipelineService pipelineSvc.Interface
	TriggerService  *triggerSvc.Service
	Authenticator   authn.Authenticator
	EventSystem     *events.System
	EventContext    *provider.EventContext
	JobScheduler    *jobsSvc.Scheduler
	Triggerer       triggerer.Triggerer
	Poller           *poller.Poller
	ScanRuleSvc      scanSvc.IRule
	ScanTaskSvc      scanSvc.IScanTask
	ScanPlanSvc      scanSvc.IScanPlan
	ScanPlanStore    scanStore.IScanPlan
	ScanTriggerStore scanStore.IScanTrigger
	RepoStore        repoStore.Interface
	RepoService      repoSvc.Interface
	TokenStore       tokenStore.Interface
	ArtifactService  artifactSvc.Interface
	ArtifactAuthorizer authorizer.Authorizer
}

func NewApp(mode string) (*App, func(), error) {
	panic(wire.Build(
		provider.ConfigWireSet,
		provider.LogWireSet,
		provider.DBWireSet,
		provider.PrincipalWireSet,
		provider.GitWireSet,
		provider.SpaceWireSet,
		provider.RepoWireSet,
		provider.BranchWireSet,
		provider.TagWireSet,
		provider.PipelineWireSet,
		provider.ExecutionWireSet,
		provider.RunnerWireSet,
		provider.DockerRunnerWireSet,
		provider.ScanRuleSetWireSet,
		provider.ScanRuleWireSet,
		provider.FitWireSet,
		provider.CommitWireSet,
		provider.WebhookWireSet,
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
