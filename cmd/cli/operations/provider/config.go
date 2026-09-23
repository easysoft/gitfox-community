// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/pkg/config"
)

type ModeName string

var ConfigWireSet = wire.NewSet(NewSystemConfig, NewSystemMode, NewDevopsConfig, NewCiConfig, NewURLConfig, NewServerConfig)

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

func NewServerConfig(cfg *config.Config) config.ServerConfig {
	return cfg.Server
}
