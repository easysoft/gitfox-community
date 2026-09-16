package provider

import (
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
	"github.com/google/wire"
)

var URLWireSet = wire.NewSet(
	NewURLProvider,
)

func NewURLProvider(cfg *config.Config) (url.Provider, error) {
	return url.NewProvider(
		cfg.URL.Internal,
		cfg.URL.Container,
		cfg.URL.API,
		cfg.URL.Git,
		cfg.URL.GitSSH,
		cfg.SSH.DefaultUser,
		cfg.SSH.Enable,
		cfg.URL.UI,
		cfg.URL.Registry,
	)
}
