// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
