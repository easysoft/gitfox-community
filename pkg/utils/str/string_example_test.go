// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package str

import (
	"fmt"
	"testing"
)

func TestSplitEx(t *testing.T) {
	result1 := SplitEx(" a b c ", "", true)

	result2 := SplitEx(" a b c ", " ", false)
	result3 := SplitEx(" a b c ", " ", true)

	result4 := SplitEx("a = b = c = ", " = ", false)
	result5 := SplitEx("a = b = c = ", " = ", true)
	result6 := SplitEx("dev/,main/,test/", ",", true)

	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)
	fmt.Println(result4)
	fmt.Println(result5)
	fmt.Println(result6)

	// []
	// [ a b c ]
	// [a b c]
	// [a b c ]
	// [a b c]
	// [dev/ main/ test/]
}
