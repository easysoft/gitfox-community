package provider

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/google/wire"
)

type SystemLog *logger.SugaredMateLogger
type APILog *logger.SugaredMateLogger

var LogWireSet = wire.NewSet(
	NewSystemLogger,
	NewAPILogger,
)

func NewSystemLogger(config *config.Config) SystemLog {
	return logger.NewZapMateLogger(config.Logging, logger.LogTypeSystem).SetAsyncer(1000).Sugar()
}

func NewAPILogger(config *config.Config) APILog {
	return logger.NewZapMateLogger(config.Logging, logger.LogTypeAPIServer).SetAsyncer(1000).Sugar()
}
