// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

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
