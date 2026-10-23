// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)

package importer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSVNCloneURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		slug    string
		want    string
		wantErr bool
	}{
		{
			name: "normal",
			host: "https://repo.oop.cc/svn/repo",
			slug: "zentaopms",
			want: "https://repo.oop.cc/svn/repo/zentaopms",
		},
		{
			name: "double slash in host",
			host: "https://repo.oop.cc//svn/repo",
			slug: "zentaopms",
			want: "https://repo.oop.cc/svn/repo/zentaopms",
		},
		{
			name: "trailing and leading slash",
			host: "https://repo.oop.cc/svn/repo/",
			slug: "/zentaopms",
			want: "https://repo.oop.cc/svn/repo/zentaopms",
		},
		{
			name: "duplicate slashes in slug",
			host: "https://repo.oop.cc/svn/repo",
			slug: "//zentaopms//sub/",
			want: "https://repo.oop.cc/svn/repo/zentaopms/sub",
		},
		{
			name: "empty slug returns repository root",
			host: "https://repo.oop.cc/svn/repo/",
			slug: "",
			want: "https://repo.oop.cc/svn/repo",
		},
		{
			name: "host without path",
			host: "https://repo.oop.cc",
			slug: "svn/repo/zentaopms",
			want: "https://repo.oop.cc/svn/repo/zentaopms",
		},
		{
			name: "svn+ssh",
			host: "svn+ssh://user@repo.oop.cc/svn/repo",
			slug: "zentaopms",
			want: "svn+ssh://user@repo.oop.cc/svn/repo/zentaopms",
		},
		{
			// legacy provider data: "svn:/myz.oop.cc//test" was split at the
			// first "//" into host "svn:/" and slug "myz.oop.cc//test".
			name: "legacy host missing authority separator",
			host: "svn:/",
			slug: "myz.oop.cc//test",
			want: "svn://myz.oop.cc//test",
		},
		{
			name: "legacy host-only scheme",
			host: "svn://",
			slug: "myz.oop.cc//test",
			want: "svn://myz.oop.cc//test",
		},
		{
			name: "legacy host with single slash and path, empty slug",
			host: "svn:/myz.oop.cc/test",
			slug: "",
			want: "svn://myz.oop.cc/test",
		},
		{
			name: "legacy host with single slash and path, extra slug",
			host: "svn:/myz.oop.cc",
			slug: "test",
			want: "svn://myz.oop.cc/test",
		},
		{
			// provider.url is host-only, while connector slug still repeats
			// the host: "svn://myz.oop.cc/" + "myz.oop.cc//test".
			name: "legacy host duplicated in slug",
			host: "svn://myz.oop.cc/",
			slug: "myz.oop.cc//test",
			want: "svn://myz.oop.cc/test",
		},
		{
			name: "legacy host duplicated in slug with provider path",
			host: "svn://myz.oop.cc/svn",
			slug: "myz.oop.cc//test",
			want: "svn://myz.oop.cc/svn/test",
		},
		{
			name: "legacy host duplicated in slug case-insensitive",
			host: "svn://Myz.OOP.cc/",
			slug: "myz.oop.cc/test",
			want: "svn://Myz.OOP.cc/test",
		},
		{
			name:    "invalid url",
			host:    "https://repo.oop.cc/svn/repo\x7f",
			slug:    "zentaopms",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildSVNCloneURL(tt.host, tt.slug)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
