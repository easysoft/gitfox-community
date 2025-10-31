//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/provider"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
}

func NewApp(mode string) (*App, func(), error) {
	panic(wire.Build(
		provider.ConfigWireSet,
		provider.LogWireSet,
		provider.DBWireSet,
		provider.SpaceWireSet,
		provider.ControllersWireSet,
		provider.I18nWireSet,
		wire.Struct(new(App), "*")),
	)
}
