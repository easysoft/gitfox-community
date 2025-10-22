package main

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cli"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cli/server"

	"github.com/quicklyon/kingpin/v2"
)

const (
	application = "gitfox"
	description = "Gitfox"
)

func main() {
	// 解析命令行参数
	args := cli.GetArguments()
	app := kingpin.New(application, description)
	server.Register(app)
	kingpin.MustParse(app.Parse(args))
}
