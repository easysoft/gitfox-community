package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/git"
	publickeyService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/ssh"
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
