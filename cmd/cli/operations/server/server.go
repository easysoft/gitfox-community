package server

import (
	"fmt"

	"github.com/quicklyon/kingpin/v2"
)

type command struct {
	mode string
}

func (c *command) run(*kingpin.ParseContext) error {
	app, cleanup, err := NewAppServer(c.mode)
	if err != nil {
		return fmt.Errorf("failed to initialize app server: %w", err)
	}
	defer cleanup()

	// TODO: 数据清理/整合
	// TODO: 启动 http server
	fmt.Printf("%s:%d\n", app.SystemConfig.Server.Host, app.SystemConfig.Server.Port)
	// srv := &http.Server{
	// 	Addr:    fmt.Sprintf("%s:%d", app.SystemConfig.Server.Host, app.SystemConfig.Server.Port),
	// 	Handler: router.NewRouter(c.mode, app.APILog),
	// }
	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 	}
	// }()
	// TODO: 优雅关停

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
