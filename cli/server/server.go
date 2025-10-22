package server

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/quicklyon/kingpin/v2"
)

type command struct {
	mode string
}

func (c *command) run(*kingpin.ParseContext) error {
	var err error
	config, err := config.LoadConfig(c.mode)
	logger := logger.NewZapMateLogger(config.Logging, logger.LogTypeSystem).SetAsyncer(1000).Sugar()
	logger.Info("调试日志内容")
	if err != nil {
		return errors.Wrap(err, "加载配置失败")
	}
	fmt.Println(config)

	return err
}

// Register the server command.
func Register(app *kingpin.Application) {
	c := new(command)

	cmd := app.Command("server", "starts the server").
		Action(c.run)

	cmd.Arg("mode", "set the server mode").
		Default("dev").
		StringVar(&c.mode)
}
