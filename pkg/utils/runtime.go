// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package utils

import (
	"runtime"
	"strings"
)

func GetCallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "Unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "Unknown"
	}

	fullName := fn.Name()
	if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
		return fullName[lastDot+1:]
	}

	return fullName
}
