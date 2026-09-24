// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package git

import "github.com/easysoft/gitfox-community/pkg/errors"

// ReadParams contains the base parameters for read operations.
// 读取操作的基本参数
type ReadParams struct {
	RepoUID             string
	AlternateObjectDirs []string
}

func (p ReadParams) Validate() error {
	if p.RepoUID == "" {
		return errors.NotFound.New("repository id cannot be empty")
	}
	return nil
}
