package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

type ModeName string

var ConfigWireSet = wire.NewSet(NewSystemConfig, NewSystemMode, NewDevopsConfig, NewCiConfig, NewURLConfig)

func NewSystemMode(mode string) ModeName {
	return ModeName(mode)
}

func NewSystemConfig(mode ModeName) (*config.Config, error) {
	cfg, err := config.LoadConfig(string(mode))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewDevopsConfig(cfg *config.Config) config.DevopsConfig {
	return cfg.Devops
}

func NewCiConfig(cfg *config.Config) config.CiConfig {
	return cfg.Ci
}

func NewURLConfig(cfg *config.Config) config.URL {
	return cfg.URL
}
