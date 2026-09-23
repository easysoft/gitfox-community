// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package svn

import (
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/google/wire"
)

// ProvideSVN returns a new SVN Service implementing Interface.
func ProvideSVN(cfg *config.Config) (Interface, error) {
	return New(cfg)
}

// SvnWireSet provides the SVN service binding.
var SvnWireSet = wire.NewSet(
	ProvideSVN,
)
