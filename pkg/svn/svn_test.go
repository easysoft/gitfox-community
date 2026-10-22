// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"bytes"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSplitAuthorityFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.pem")

	var bundle []byte
	bundle = append(bundle, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("cert-one")})...)
	bundle = append(bundle, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("cert-two")})...)
	// duplicate certificates should be emitted only once
	bundle = append(bundle, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("cert-one")})...)
	if err := os.WriteFile(src, bundle, 0o600); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	got, err := splitAuthorityFile(src, filepath.Join(dir, "ca"))
	if err != nil {
		t.Fatalf("splitAuthorityFile() error = %v", err)
	}

	paths := strings.Split(got, ";")
	if len(paths) != 2 {
		t.Fatalf("splitAuthorityFile() returned %d files, want 2: %v", len(paths), paths)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read split file %q: %v", p, err)
		}
		if n := bytes.Count(data, []byte("BEGIN CERTIFICATE")); n != 1 {
			t.Fatalf("split file %q contains %d certificates, want 1", p, n)
		}
	}

	if runtime.GOOS != "windows" {
		dirInfo, err := os.Stat(filepath.Join(dir, "ca"))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

		for _, p := range paths {
			info, err := os.Stat(p)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		}
	}
}

func TestSplitAuthorityFileRejectsSymlinkDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on Windows")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("cert-one")})
	require.NoError(t, os.WriteFile(src, cert, 0o600))

	realDir := filepath.Join(dir, "real-ca")
	require.NoError(t, os.Mkdir(realDir, 0o700))
	linkDir := filepath.Join(dir, "ca-link")
	require.NoError(t, os.Symlink(realDir, linkDir))

	_, err := splitAuthorityFile(src, linkDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be a symlink")
}

func TestNormalizeTrustServerCertFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "single", raw: "unknown-ca", want: "unknown-ca"},
		{name: "spaces and duplicates", raw: " Unknown-CA , expired,unknown-ca ", want: "unknown-ca,expired"},
		{
			name: "all allowed",
			raw:  "unknown-ca,cn-mismatch,expired,not-yet-valid,other",
			want: "unknown-ca,cn-mismatch,expired,not-yet-valid,other",
		},
		{name: "invalid", raw: "unknown-ca,everything", wantErr: true},
		{name: "empty", raw: " , ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeTrustServerCertFailures(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewRejectsAuthorityFileWithTrustFailures(t *testing.T) {
	oldAuthorityFiles, oldFailures := command.SSLAuthorityFiles, command.TrustServerCertFailures
	t.Cleanup(func() {
		command.SSLAuthorityFiles, command.TrustServerCertFailures = oldAuthorityFiles, oldFailures
	})

	_, err := New(&config.Config{SVN: config.SvnConfig{
		SSLAuthorityFile:        "/tmp/ca.pem",
		TrustServerCertFailures: "unknown-ca",
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be configured together")
}

func TestSplitAuthorityFileNoCertificate(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty.pem")
	if err := os.WriteFile(src, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := splitAuthorityFile(src, filepath.Join(dir, "ca")); err == nil {
		t.Fatal("splitAuthorityFile() expected error for file without certificates")
	}
}
