// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:   "standard linux output",
			output: "svn, version 1.14.3 (r1914484)\n   compiled Apr  1 2024, 04:55:49 on x86_64-pc-linux-gnu\n",
			want:   "1.14.3",
		},
		{
			name:   "older version format",
			output: "svn, version 1.8.10 (r1615264)\n   compiled Aug  8 2014, 10:40:21 on x86_64-unknown-linux-gnu\n",
			want:   "1.8.10",
		},
		{
			name:   "version with prerelease",
			output: "svn, version 1.15.0-dev\n   compiled Jan 10 2025, 10:00:00 on x86_64-linux\n",
			want:   "1.15.0-dev",
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name:    "malformed output",
			output:  "svn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseVersion() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parseVersion() unexpected error: %v", err)
				return
			}
			if got.Original() != tt.want {
				t.Errorf("parseVersion() = %q, want %q", got.Original(), tt.want)
			}
		})
	}
}
