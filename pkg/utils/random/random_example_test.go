// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package random

import (
	"fmt"
	"regexp"
)

func ExampleRandInt() {
	result := RandInt(1, 10)

	if result >= 1 && result < 10 {
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleRandString() {
	pattern := `^[a-zA-Z]+$`
	reg := regexp.MustCompile(pattern)

	s := RandString(6)

	result1 := reg.MatchString(s)
	result2 := len(s)

	fmt.Println(result1)
	fmt.Println(result2)

	// Output:
	// true
	// 6
}
