package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/git"
	publickeyService "github.com/easysoft/gitfox-community/app/service/publickey"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/ssh"
)

var SSHWireSet = wire.NewSet(
	ProvideSSHServer,
)

func ProvideSSHServer(
	cfg *config.Config,
	verifier publickeyService.Interface,
	GitCtrl *git.GitController,
) *ssh.Server {
	return &ssh.Server{
		Host:              cfg.SSH.Host,
		Port:              cfg.SSH.Port,
		DefaultUser:       cfg.SSH.DefaultUser,
		ServerHostKeysDir: cfg.SSH.HostKeysDir,
		Verifier:          verifier,
		GitCtrl:           GitCtrl,
	}
}
