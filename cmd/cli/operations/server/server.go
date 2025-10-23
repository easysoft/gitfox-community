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
	// 先加载配置
	cfg, err := config.LoadConfig(c.mode)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败,%v", err))
	}

	// 使用加载的配置初始化 logger
	log := logger.NewZapMateLogger(cfg.Logging, logger.LogTypeSystem).SetAsyncer(1000).Sugar()
	defer log.Flush() // 确保异步日志在程序退出前写入

	log.Info("调试日志内容")
	fmt.Println(cfg)

	return nil
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
