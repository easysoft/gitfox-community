package config

import "testing"

func TestLoadDebugConfig(t *testing.T) {
	LoadConfig("debug")
}


func TestBackfillURLsRepoPortExposure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cfg           *Config
		wantUI        string
		wantGitURL    string
		wantGitSSHURL string
	}{
		{
			name: "custom ports exposed when enabled",
			cfg: &Config{
				URL: URL{
					ExposeRepoPort: true,
				},
				Server: ServerConfig{
					Proto:      schemeHTTP,
					ServerHost: "gitfox.example.com",
					Port:       3000,
				},
				SSH: SSHConfig{
					Enable: true,
					Host:   "0.0.0.0",
					Port:   3322,
				},
			},
			wantUI:        "http://gitfox.example.com:3000",
			wantGitURL:    "http://gitfox.example.com:3000/git",
			wantGitSSHURL: "ssh://gitfox.example.com:3322",
		},
		{
			name: "ports hidden when disabled",
			cfg: &Config{
				URL: URL{
					ExposeRepoPort: false,
				},
				Server: ServerConfig{
					Proto:      schemeHTTP,
					ServerHost: "gitfox.example.com",
					Port:       3000,
				},
				SSH: SSHConfig{
					Enable: true,
					Host:   "0.0.0.0",
					Port:   3322,
				},
			},
			wantUI:        "http://gitfox.example.com",
			wantGitURL:    "http://gitfox.example.com/git",
			wantGitSSHURL: "ssh://gitfox.example.com",
		},
		{
			name: "default http https and ssh ports stay hidden",
			cfg: &Config{
				URL: URL{
					ExposeRepoPort: true,
					Base:           "https://gitfox.example.com:443/base",
				},
				SSH: SSHConfig{
					Enable: true,
					Host:   "gitfox.example.com",
					Port:   22,
				},
			},
			wantUI:        "https://gitfox.example.com/base",
			wantGitURL:    "https://gitfox.example.com/base/git",
			wantGitSSHURL: "ssh://gitfox.example.com",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := backfillURLs(tt.cfg); err != nil {
				t.Fatalf("backfillURLs() error = %v", err)
			}

			if tt.cfg.URL.UI != tt.wantUI {
				t.Fatalf("UI URL = %q, want %q", tt.cfg.URL.UI, tt.wantUI)
			}
			if tt.cfg.URL.Git != tt.wantGitURL {
				t.Fatalf("Git URL = %q, want %q", tt.cfg.URL.Git, tt.wantGitURL)
			}
			if tt.cfg.URL.GitSSH != tt.wantGitSSHURL {
				t.Fatalf("GitSSH URL = %q, want %q", tt.cfg.URL.GitSSH, tt.wantGitSSHURL)
			}
		})
	}
}