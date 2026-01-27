//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/ssh"
	"gorm.io/gorm"
)

type App struct {
	SystemConfig    *config.Config
	SystemLog       *provider.SystemLog
	APILog          *provider.APILog
	DB              *gorm.DB
	SqliteDB        *database.SqliteDB
	Controller      *router.Controller
	I18nManager     *provider.I18nManager
	GithookCtrl     *githook.Controller
	SSHServer       *ssh.Server
	UserService     userSvc.Interface
	PipelineService pipelineSvc.Interface
	Authenticator   authn.Authenticator
	EventSystem     *events.System
	EventContext    *provider.EventContext
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
		provider.ScanRuleSetWireSet,
		provider.ScanRuleWireSet,
		provider.CommitWireSet,
		provider.ControllersWireSet,
		provider.URLWireSet,
		provider.I18nWireSet,
		provider.SSHWireSet,
		provider.LockWireSet,
		provider.EventWireSet,
		provider.GitEventWireSet,
		provider.PullReqEventWireSet,
		provider.PubSubProvider,
		provider.AuthWireSet, wire.Struct(new(App), "*")),
	)
}
