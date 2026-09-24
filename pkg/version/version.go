// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package version

import (
	"fmt"
)

var (
	Version       string
	BuildDate     string
	GitCommitHash string
	GitBranch     string
)

// GetVersionWithHash returns the version string with the git commit hash and build date.
func GetVersionWithHash() string {
	return fmt.Sprintf("%s-%s-%s", Version, GitCommitHash, BuildDate)
}
