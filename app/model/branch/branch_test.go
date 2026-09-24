// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package branch

import (
	"strings"
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// 合法
		{"plain", "dev", false},
		{"with_prefix", "feature/login", false},
		{"with_dash_inner", "feature-x", false},
		{"with_underscore", "feature_x", false},
		{"with_dot_inner", "v1.0", false},

		// 长度
		{"too_long", strings.Repeat("a", 256), true},
		{"max_length", "a" + strings.Repeat("b", 254), false},

		// 开头
		{"leading_dot", ".hidden", true},
		{"leading_hyphen", "-opt", true},
		{"leading_digit", "1abc", true},

		// 结尾
		{"trailing_slash", "feature/", true},
		{"trailing_dot", "feature.", true},
		{"trailing_lock", "feature.lock", true},

		// 连续序列
		{"dot_dot", "feat..x", true},
		{"double_slash", "feat//x", true},

		// 禁用字符
		{"space", "feat x", true},
		{"tab", "feat\tx", true},
		{"newline", "feat\nx", true},
		{"tilde", "feat~x", true},
		{"caret", "feat^x", true},
		{"colon", "feat:x", true},
		{"question", "feat?x", true},
		{"asterisk", "feat*x", true},
		{"open_bracket", "feat[x", true},
		{"backslash", "feat\\x", true},
		{"open_brace", "feat{x", true},
		{"close_brace", "feat}x", true},
		{"at", "feat@x", true},
		{"bang", "feat!x", true},
		{"dollar", "feat$x", true},
		{"amp", "feat&x", true},
		{"paren_open", "feat(x", true},
		{"paren_close", "feat)x", true},
		{"equal", "feat=x", true},
		{"plus", "feat+x", true},
		{"pipe", "feat|x", true},
		{"lt", "feat<x", true},
		{"gt", "feat>x", true},
		{"semi", "feat;x", true},
		{"dquote", "feat\"x", true},
		{"squote", "feat'x", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateBranchName(c.input)
			if (err != nil) != c.wantErr {
				t.Fatalf("validateBranchName(%q) err=%v, wantErr=%v", c.input, err, c.wantErr)
			}
		})
	}
}
