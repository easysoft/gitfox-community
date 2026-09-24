// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package rbac

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWildcardMatch(t *testing.T) {
	var tests = []struct {
		pattern string
		target  string
		result  bool
	}{
		{"*", "1", true},
		{"*", "f2", true},
		{"/tmp/*", "/tmp/file1", true},
		{"/tmp/*", "/tmp/d1/file1", false},
		{"/tmp/**", "/tmp/d1/file1", true},
	}

	for _, test := range tests {
		ok, err := WildcardMatchFunc(test.target, test.pattern)
		require.NoError(t, err, test.pattern, test.target)
		require.Equal(t, test.result, ok, "%s %s", test.pattern, test.target)
	}
}
