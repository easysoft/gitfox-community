// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package server

import (
	"github.com/drone/runner-go/poller"
	"gorm.io/gorm"

	"github.com/easysoft/gitfox-community/app/api/controller/githook"
	"github.com/easysoft/gitfox-community/app/api/router"
	ctnHandler "github.com/easysoft/gitfox-community/app/artifact/handler"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	artifactSvc "github.com/easysoft/gitfox-community/app/service/artifact"
	importerSvc "github.com/easysoft/gitfox-community/app/service/importer"
	jobsSvc "github.com/easysoft/gitfox-community/app/service/jobs"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	scanSvc "github.com/easysoft/gitfox-community/app/service/scan"
	triggerSvc "github.com/easysoft/gitfox-community/app/service/trigger"
	userSvc "github.com/easysoft/gitfox-community/app/service/user"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	tokenStore "github.com/easysoft/gitfox-community/app/store/token"
	"github.com/easysoft/gitfox-community/cmd/cli/operations/provider"
	"github.com/easysoft/gitfox-community/pkg/auth/authn"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/events"
	gitPkg "github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/ssh"
	svnPkg "github.com/easysoft/gitfox-community/pkg/svn"
)

// App is the fully assembled application dependency graph.
//
// It is intentionally declared in a plain (non wireinject) file so that both
// the community wire injector (NewApp) and downstream overlay injectors can
// reference the same type without triggering duplicate declarations in
// generated code.
type App struct {
	SystemConfig     *config.Config
	SystemLog        *provider.SystemLog
	APILog           *provider.APILog
	DB               *gorm.DB
	Controller       *router.Controller
	I18nManager      *provider.I18nManager
	GithookCtrl      *githook.Controller
	SSHServer        *ssh.Server
	UserService      userSvc.Interface
	PipelineService  pipelineSvc.Interface
	TriggerService   *triggerSvc.Service
	Authenticator    authn.Authenticator
	EventSystem      *events.System
	EventContext     *provider.EventContext
	JobScheduler     *jobsSvc.Scheduler
	Triggerer        triggerer.Triggerer
	Poller           *poller.Poller
	ScanRuleSvc      scanSvc.IRule
	ScanTaskSvc      scanSvc.IScanTask
	ScanPlanSvc      scanSvc.IScanPlan
	ScanPlanStore    scanStore.IScanPlan
	ScanTriggerStore scanStore.IScanTrigger
	RepoStore        repoStore.Interface
	RepoService      repoSvc.Interface
	RepoImporterSvc  *importerSvc.Repository
	TokenStore       tokenStore.Interface
	ArtifactService  artifactSvc.Interface
	ContainerHandler *ctnHandler.ContainerHandler
	GitService       gitPkg.Interface
	SvnService       svnPkg.Interface
}
