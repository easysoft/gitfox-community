package artifact

import "testing"

func TestCreateArtifactViewReq_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateArtifactViewReq
		wantErr bool
	}{
		{name: "generic 允许中文", req: CreateArtifactViewReq{Name: "测试库", Format: "generic"}, wantErr: false},
		{name: "raw 允许中文", req: CreateArtifactViewReq{Name: "中文_raw", Format: "raw"}, wantErr: false},
		{name: "container 禁中文", req: CreateArtifactViewReq{Name: "测试container", Format: "container"}, wantErr: true},
		{name: "helm 禁中文", req: CreateArtifactViewReq{Name: "中文helm", Format: "helm"}, wantErr: true},
		{name: "maven 禁中文", req: CreateArtifactViewReq{Name: "中文maven", Format: "maven"}, wantErr: true},
		{name: "pypi 禁中文", req: CreateArtifactViewReq{Name: "中文pypi", Format: "pypi"}, wantErr: true},
		{name: "npm 禁中文", req: CreateArtifactViewReq{Name: "中文npm", Format: "npm"}, wantErr: true},
		{name: "container 仅 ASCII", req: CreateArtifactViewReq{Name: "my-container_01", Format: "container"}, wantErr: false},
		{name: "空名称", req: CreateArtifactViewReq{Name: "   ", Format: "generic"}, wantErr: true},
		{name: "非法字符 /", req: CreateArtifactViewReq{Name: "bad/name", Format: "generic"}, wantErr: true},
		{name: "非法字符 .", req: CreateArtifactViewReq{Name: "bad.name", Format: "generic"}, wantErr: true},
		{name: "非法字符 空白", req: CreateArtifactViewReq{Name: "bad name", Format: "generic"}, wantErr: true},
		{name: "不支持的 format", req: CreateArtifactViewReq{Name: "ok", Format: "rpm"}, wantErr: true},
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
		{name: "format 为空（未回填）", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: ""}, wantErr: true},
		{name: "format 不合法", req: CreateGroupReq{ArtifactID: 1, Names: "a", Format: "rpm"}, wantErr: true},
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
