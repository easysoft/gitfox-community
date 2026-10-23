// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestPreparePluginsZipURLDefault(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantContent string
		wantChange  bool
	}{
		{
			name:        "https assignment is commented",
			content:     "A=1\nGITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\nB=2\n",
			wantContent: "A=1\n# GITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\nB=2\n",
			wantChange:  true,
		},
		{
			name:        "uppercase http assignment is commented",
			content:     "GITFOX_CI_PLUGINS_ZIP_URL=HTTP://old.example.com/plugin.tar.gz\n",
			wantContent: "# GITFOX_CI_PLUGINS_ZIP_URL=HTTP://old.example.com/plugin.tar.gz\n",
			wantChange:  true,
		},
		{
			name:        "spaces and export are supported",
			content:     "  export GITFOX_CI_PLUGINS_ZIP_URL = \"http://old.example.com/a.tar.gz\" # old\n",
			wantContent: "  # export GITFOX_CI_PLUGINS_ZIP_URL = \"http://old.example.com/a.tar.gz\" # old\n",
			wantChange:  true,
		},
		{
			name:        "CRLF is preserved",
			content:     "A=1\r\nGITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\r\nB=2\r\n",
			wantContent: "A=1\r\n# GITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\r\nB=2\r\n",
			wantChange:  true,
		},
		{
			name: "duplicate active assignments are all commented",
			content: "GITFOX_CI_PLUGINS_ZIP_URL=http://first.example.com/plugin.tar.gz\n" +
				"GITFOX_CI_PLUGINS_ZIP_URL=https://last.example.com/plugin.tar.gz\n",
			wantContent: "# GITFOX_CI_PLUGINS_ZIP_URL=http://first.example.com/plugin.tar.gz\n" +
				"# GITFOX_CI_PLUGINS_ZIP_URL=https://last.example.com/plugin.tar.gz\n",
			wantChange: true,
		},
		{
			name:        "local path is unchanged",
			content:     "GITFOX_CI_PLUGINS_ZIP_URL=/app/ciPlugins/plugins.tar.gz\n",
			wantContent: "GITFOX_CI_PLUGINS_ZIP_URL=/app/ciPlugins/plugins.tar.gz\n",
		},
		{
			name:        "commented assignment is unchanged",
			content:     "# GITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\n",
			wantContent: "# GITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\n",
		},
		{
			name:        "similar key is unchanged",
			content:     "GITFOX_CI_PLUGINS_ZIP_URL_EXTRA=https://old.example.com/plugin.tar.gz\n",
			wantContent: "GITFOX_CI_PLUGINS_ZIP_URL_EXTRA=https://old.example.com/plugin.tar.gz\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			envPath := filepath.Join(workDir, ProdEnvFileName)
			require.NoError(t, os.WriteFile(envPath, []byte(tt.content), 0644))

			change, err := preparePluginsZipURLDefault(workDir)
			require.NoError(t, err)
			if !tt.wantChange {
				require.Nil(t, change)
				actual, err := os.ReadFile(envPath)
				require.NoError(t, err)
				require.Equal(t, tt.wantContent, string(actual))
				return
			}

			require.NotNil(t, change)
			require.Equal(t, tt.wantContent, string(change.updated))
			envMap, err := godotenv.Unmarshal(tt.wantContent)
			require.NoError(t, err)
			_, exists := envMap[pluginsZipURLEnvKey]
			require.False(t, exists)
		})
	}
}

func TestPluginsZipURLChangeApplyAndRestore(t *testing.T) {
	workDir := t.TempDir()
	envPath := filepath.Join(workDir, ProdEnvFileName)
	original := "GITFOX_SERVER_HOST=gitfox.example.com\n" +
		"GITFOX_CI_PLUGINS_ZIP_URL=https://old.example.com/plugin.tar.gz\n"
	require.NoError(t, os.WriteFile(envPath, []byte(original), 0600))

	change, err := preparePluginsZipURLDefault(workDir)
	require.NoError(t, err)
	require.NotNil(t, change)

	require.NoError(t, change.apply())
	updated, err := os.ReadFile(envPath)
	require.NoError(t, err)
	require.Equal(t, change.updated, updated)
	require.NotEmpty(t, change.backupPath)
	require.NotEqual(t, envPath, change.backupPath)
	require.Contains(t, filepath.Base(change.backupPath), ProdEnvFileName+".upgrade-bak-")
	backup, err := os.ReadFile(change.backupPath)
	require.NoError(t, err)
	require.Equal(t, original, string(backup))

	envMap, err := godotenv.Unmarshal(string(updated))
	require.NoError(t, err)
	require.NotContains(t, envMap, pluginsZipURLEnvKey)
	require.Equal(t, "gitfox.example.com", envMap["GITFOX_SERVER_HOST"])

	require.NoError(t, change.restore())
	restored, err := os.ReadFile(envPath)
	require.NoError(t, err)
	require.Equal(t, original, string(restored))
	backupAfterRestore, err := os.ReadFile(change.backupPath)
	require.NoError(t, err)
	require.Equal(t, original, string(backupAfterRestore))
}

func TestPreparePluginsZipURLDefaultMissingEnv(t *testing.T) {
	_, err := preparePluginsZipURLDefault(t.TempDir())
	require.Error(t, err)
	require.ErrorContains(t, err, ProdEnvFileName)
}
