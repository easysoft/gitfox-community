//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type App struct {
	SystemConfig *config.Config
	SystemLog    *provider.SystemLog
	APILog       *provider.APILog
	MysqlDB      *database.MysqlDB
	SqliteDB     *database.SqliteDB
	Controller   *router.Controller
}

func NewApp(mode string) (*App, func(), error) {
	panic(wire.Build(
		provider.ConfigWireSet,
		provider.LogWireSet,
		provider.DBWireSet,
		provider.SpaceWireSet,
		provider.ControllersWireSet,
		wire.Struct(new(App), "*")),
	)
}
