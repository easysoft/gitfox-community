package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

type ModeName string

var ConfigWireSet = wire.NewSet(NewSystemConfig, NewSystemMode)

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
