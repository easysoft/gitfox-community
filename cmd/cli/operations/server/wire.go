//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/ssh"
	"gorm.io/gorm"
)

type App struct {
	SystemConfig *config.Config
	SystemLog    *provider.SystemLog
	APILog       *provider.APILog
	DB           *gorm.DB
	SqliteDB     *database.SqliteDB
	Controller   *router.Controller
	I18nManager  *provider.I18nManager
	GithookCtrl  *githook.Controller
	SSHServer    *ssh.Server
	UserService  userSvc.Interface
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
		provider.ScanRuleSetWireSet,
		provider.ScanRuleWireSet,
		provider.CommitWireSet,
		provider.ControllersWireSet,
		provider.URLWireSet,
		provider.I18nWireSet,
		provider.SSHWireSet,
		wire.Struct(new(App), "*")),
	)
}
