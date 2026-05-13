package artifact

import (
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func TestCreateArtifactViewReq_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateArtifactViewReq
		wantErr bool
		wantKey i18n.I18nKey // 为空表示不校验
	}{
		{name: "generic 允许中文", req: CreateArtifactViewReq{Name: "测试库", Format: "generic"}, wantErr: false},
		{name: "raw 允许中文", req: CreateArtifactViewReq{Name: "中文_raw", Format: "raw"}, wantErr: false},
		{name: "container 禁中文", req: CreateArtifactViewReq{Name: "测试container", Format: "container"}, wantErr: true, wantKey: i18n.ArtifactViewNameNonAsciiNotAllowed},
		{name: "helm 禁中文", req: CreateArtifactViewReq{Name: "中文helm", Format: "helm"}, wantErr: true, wantKey: i18n.ArtifactViewNameNonAsciiNotAllowed},
		{name: "maven 禁中文", req: CreateArtifactViewReq{Name: "中文maven", Format: "maven"}, wantErr: true, wantKey: i18n.ArtifactViewNameNonAsciiNotAllowed},
		{name: "pypi 禁中文", req: CreateArtifactViewReq{Name: "中文pypi", Format: "pypi"}, wantErr: true, wantKey: i18n.ArtifactViewNameNonAsciiNotAllowed},
		{name: "npm 禁中文", req: CreateArtifactViewReq{Name: "中文npm", Format: "npm"}, wantErr: true, wantKey: i18n.ArtifactViewNameNonAsciiNotAllowed},
		{name: "container 仅 ASCII", req: CreateArtifactViewReq{Name: "my-container_01", Format: "container"}, wantErr: false},
		{name: "空名称", req: CreateArtifactViewReq{Name: "   ", Format: "generic"}, wantErr: true, wantKey: i18n.ArtifactViewNameRequired},
		{name: "非法字符 /", req: CreateArtifactViewReq{Name: "bad/name", Format: "generic"}, wantErr: true, wantKey: i18n.ArtifactViewNameInvalidChars},
		{name: "非法字符 .", req: CreateArtifactViewReq{Name: "bad.name", Format: "generic"}, wantErr: true, wantKey: i18n.ArtifactViewNameInvalidChars},
		{name: "非法字符 空白", req: CreateArtifactViewReq{Name: "bad name", Format: "generic"}, wantErr: true, wantKey: i18n.ArtifactViewNameInvalidChars},
		{name: "不支持的 format", req: CreateArtifactViewReq{Name: "ok", Format: "rpm"}, wantErr: true, wantKey: i18n.ArtifactViewFormatInvalid},
		{name: "空 format", req: CreateArtifactViewReq{Name: "ok", Format: ""}, wantErr: true, wantKey: i18n.ArtifactViewFormatRequired},
		{name: "空白 format", req: CreateArtifactViewReq{Name: "ok", Format: "   "}, wantErr: true, wantKey: i18n.ArtifactViewFormatRequired},
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
		{name: "generic", format: "generic", wantErr: false},
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
	}{
		{name: "artifactID 为 0", req: CreateGroupReq{ArtifactID: 0, Names: "a", Format: "generic"}, wantErr: true},
		{name: "format 为空（由服务端按 view.Format 回填，不再校验）", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: ""}, wantErr: false},
		{name: "format 未知但未启用校验（服务端回填后才生效）", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: "rpm"}, wantErr: false},
		{name: "names 含空段", req: CreateGroupReq{ArtifactID: 1, Names: "a..b", Format: "generic"}, wantErr: true},
		{name: "container 禁嵌套", req: CreateGroupReq{ArtifactID: 1, Names: "a.b", Format: "container"}, wantErr: true},
		{name: "container 单级合法", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: "container"}, wantErr: false},
		{name: "generic 多级合法", req: CreateGroupReq{ArtifactID: 1, Names: "a.b.c", Format: "generic"}, wantErr: false},
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
		})
	}
}
