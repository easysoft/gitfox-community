package server

import (
	"fmt"
	
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"github.com/quicklyon/kingpin/v2"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type command struct {
	mode   string
	logger *logger.ZapMateLogger
}


func (c *command) run(*kingpin.ParseContext) error {
	var err error
	config,err := config.LoadConfig(c.mode)
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

