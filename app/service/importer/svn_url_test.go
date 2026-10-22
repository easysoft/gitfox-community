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
