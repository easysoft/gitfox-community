// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package common

import "os"

// function used to expand environment variables.
var getenv = os.Getenv

// GetEnv 获取环境变量
func GetEnv(envstr string, fallback ...string) string {
	e := getenv(envstr)
	if e == "" && len(fallback) > 0 {
		e = fallback[0]
	}
	return e
}
