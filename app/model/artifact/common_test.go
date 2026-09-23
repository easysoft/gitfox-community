// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"testing"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

func TestCreateArtifactViewReq_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateArtifactLibReq
		wantErr bool
		wantKey i18n.I18nKey // 为空表示不校验
	}{
		{name: "file 允许中文英文点和大写", req: CreateArtifactLibReq{Name: "测试.Test_name-01", Format: "file"}, wantErr: false},
		{name: "raw 允许英文点", req: CreateArtifactLibReq{Name: "a.b_c-d", Format: "raw"}, wantErr: false},
		{name: "container 允许英文点", req: CreateArtifactLibReq{Name: "my.container_01", Format: "container"}, wantErr: false},
		{name: "container 禁中文", req: CreateArtifactLibReq{Name: "测试container", Format: "container"}, wantErr: true, wantKey: i18n.ArtifactLibNameNonAsciiNotAllowed},
		{name: "空名称", req: CreateArtifactLibReq{Name: "   ", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactLibNameRequired},
		{name: "非法字符 /", req: CreateArtifactLibReq{Name: "bad/name", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactLibNameInvalidChars},
		{name: "非法字符 空白", req: CreateArtifactLibReq{Name: "bad name", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactLibNameInvalidChars},
		{name: "不支持的 format", req: CreateArtifactLibReq{Name: "ok", Format: "rpm"}, wantErr: true, wantKey: i18n.ArtifactLibFormatInvalid},
		{name: "空 format", req: CreateArtifactLibReq{Name: "ok", Format: ""}, wantErr: true, wantKey: i18n.ArtifactLibFormatRequired},
		{name: "空白 format", req: CreateArtifactLibReq{Name: "ok", Format: "   "}, wantErr: true, wantKey: i18n.ArtifactLibFormatRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req
			err := req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil; req=%+v", tc.req)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v; req=%+v", err, tc.req)
			}
			if tc.wantKey != "" && err != nil {
				if got := errors.GetI18nKey(err); got != tc.wantKey {
					t.Fatalf("expected i18n key %q, got %q; req=%+v", tc.wantKey, got, tc.req)
				}
			}
		})
	}
}

func TestRefreshTreeReq_Validate(t *testing.T) {
	cases := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{name: "空 format 视为 all", format: "", wantErr: false},
		{name: "all", format: "all", wantErr: false},
		{name: "file", format: "file", wantErr: false},
		{name: "container", format: "container", wantErr: false},
		{name: "helm", format: "helm", wantErr: false},
		{name: "不支持的 format", format: "rpm", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &RefreshTreeReq{Format: tc.format}
			err := r.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil; format=%q", tc.format)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v; format=%q", err, tc.format)
			}
		})
	}
}

func TestCreateGroupReq_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateGroupReq
		wantErr bool
		wantKey i18n.I18nKey
	}{
		{name: "artifactID 为 0", req: CreateGroupReq{ArtifactID: 0, Names: "a", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupArtifactidRequired},
		{name: "format 为空（由服务端按 view.Format 回填，不再校验）", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: ""}, wantErr: false},
		{name: "format 未知但未启用校验（服务端回填后才生效）", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: "rpm"}, wantErr: false},
		{name: "段内允许英文点", req: CreateGroupReq{ArtifactID: 1, Names: "a.b.c", Format: "file"}, wantErr: false},
		{name: "names 含空段", req: CreateGroupReq{ArtifactID: 1, Names: "a//b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNamesInvalid},
		{name: "container 禁嵌套", req: CreateGroupReq{ArtifactID: 1, Names: "a/b", Format: "container"}, wantErr: true, wantKey: i18n.ArtifactGroupContainerNoNesting},
		{name: "container 单级合法", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: "container"}, wantErr: false},
		{name: "file 多级路径合法", req: CreateGroupReq{ArtifactID: 1, Names: "a/b/c", Format: "file"}, wantErr: false},
		{name: "点号属于段名的一部分", req: CreateGroupReq{ArtifactID: 1, Names: "a.b/c.d", Format: "file"}, wantErr: false},

		// —— 字符集禁用 ——
		{name: "段内含空格", req: CreateGroupReq{ArtifactID: 1, Names: "a b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含 Tab", req: CreateGroupReq{ArtifactID: 1, Names: "a\tb", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含换行", req: CreateGroupReq{ArtifactID: 1, Names: "a\nb", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含回车", req: CreateGroupReq{ArtifactID: 1, Names: "a\rb", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含全角空格", req: CreateGroupReq{ArtifactID: 1, Names: "a　b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含反斜杠", req: CreateGroupReq{ArtifactID: 1, Names: "a\\b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含冒号", req: CreateGroupReq{ArtifactID: 1, Names: "a:b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含星号", req: CreateGroupReq{ArtifactID: 1, Names: "a*b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含问号", req: CreateGroupReq{ArtifactID: 1, Names: "a?b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含双引号", req: CreateGroupReq{ArtifactID: 1, Names: "a\"b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含小于号", req: CreateGroupReq{ArtifactID: 1, Names: "a<b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含大于号", req: CreateGroupReq{ArtifactID: 1, Names: "a>b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},
		{name: "段内含竖线", req: CreateGroupReq{ArtifactID: 1, Names: "a|b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameInvalidChars},

		// —— 首尾空白 ——
		{name: "段首空格", req: CreateGroupReq{ArtifactID: 1, Names: " a", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
		{name: "段尾空格", req: CreateGroupReq{ArtifactID: 1, Names: "a ", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
		{name: "段首 Tab", req: CreateGroupReq{ArtifactID: 1, Names: "\ta", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
		{name: "段首全角空格", req: CreateGroupReq{ArtifactID: 1, Names: "　a", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
		{name: "中间段首空格", req: CreateGroupReq{ArtifactID: 1, Names: "a/ b", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
		{name: "中间段尾空格", req: CreateGroupReq{ArtifactID: 1, Names: "a/b ", Format: "file"}, wantErr: true, wantKey: i18n.ArtifactGroupNameLeadingTrailingSpace},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req
			err := req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil; req=%+v", tc.req)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v; req=%+v", err, tc.req)
			}
			if tc.wantKey != "" && err != nil {
				if got := errors.GetI18nKey(err); got != tc.wantKey {
					t.Fatalf("expected i18n key %q, got %q; req=%+v", tc.wantKey, got, tc.req)
				}
			}
		})
	}
}
