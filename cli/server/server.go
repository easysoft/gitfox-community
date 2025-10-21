package server

import (
	"fmt"
	
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"github.com/quicklyon/kingpin/v2"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

type command struct {
	mode   string
	logger *logger.ZapMateLogger
}


func (c *command) run(*kingpin.ParseContext) error {
	var err error
	config.LoadConfig(c.mode)

	fmt.Println("Starting server in mode:", c.mode)
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

