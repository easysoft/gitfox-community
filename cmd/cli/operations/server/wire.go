//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
}

func NewApp(mode string) (*App, func(), error) {
	panic(wire.Build(
		provider.ConfigWireSet,
		provider.LogWireSet,
		provider.DBWireSet,
		provider.PrincipalWireSet,
		provider.GitWireSet,
		provider.SpaceWireSet,
		provider.RepoWireSet, // 必须在 Branch/Tag/Commit 之前，因为它们依赖 repoService
		provider.BranchWireSet,
		provider.TagWireSet,
		provider.CommitWireSet,
		provider.ControllersWireSet,
		provider.URLWireSet,
		provider.I18nWireSet,
		wire.Struct(new(App), "*")),
	)
}
