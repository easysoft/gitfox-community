package server

import (
	"fmt"

	"github.com/quicklyon/kingpin/v2"
)

type command struct {
	mode     string
	envfile  string
}

// Register the server command.
func Register(app *kingpin.Application) {
	c := new(command)

	cmd := app.Command("server", "starts the server").
		Action(c.run)

	cmd.Arg("mode", "load the environment variable file").
		Default("dev").
		StringVar(&c.mode)

	// cmd.Arg("envfile", "load the environment variable file").
	// 	Default("").
	// 	StringVar(&c.envfile)
}


func (c *command) run(*kingpin.ParseContext) error {
	var err error
	fmt.Println("Starting server in mode:", c.mode)
	return err
}