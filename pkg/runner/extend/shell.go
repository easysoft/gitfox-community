// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package extend

import (
	"bytes"
	"fmt"
)

const customScript = `
test -d $(dirname "$GITFOX_CUSTOM_ENV") || mkdir -p $(dirname "$GITFOX_CUSTOM_ENV")
if [ -f "${GITFOX_CUSTOM_ENV}" ];then
	. "${GITFOX_CUSTOM_ENV}"
fi
`

func IncludeShell(buf *bytes.Buffer) {
	fmt.Fprintf(buf, customScript)
}
