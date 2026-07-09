package service

import (
	"os"
	"path/filepath"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamInit = "init"
)

func RegisterInit(app *kingpin.Application) {
	app.Command(ParamInit, clii18n.T(i18n.CliInstallInitDescription)).
		Action(runInit)
}

func WorkDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if realPath, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = realPath
	}
	return filepath.Dir(exePath), nil
}
