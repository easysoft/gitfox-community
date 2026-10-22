// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)

package command

import (
	"bytes"
	"context"
	"reflect"
	"runtime"
	"testing"
)

// TestRunInheritsEnv 验证 svn 子进程能继承代理/证书/HOME 这类关键环境变量，
// 同时保留通过 Envs 显式传入的变量。
func TestRunInheritsEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test relies on POSIX sh")
	}

	t.Setenv("HOME", "/tmp/svn-test-home")
	t.Setenv("SSL_CERT_FILE", "/tmp/svn-test-ca.pem")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:3128")

	cmd := &Command{
		Executable: "sh",
		Flags: []string{"-c",
			`printf '%s|%s|%s|%s|%s' "$HOME" "$SSL_CERT_FILE" "$HTTPS_PROXY" "$SVN_TEST_ENV" "$LANG"`},
		Envs: Envs{"SVN_TEST_ENV": "explicit"},
	}

	stdout := &bytes.Buffer{}
	if err := cmd.Run(context.Background(), WithStdout(stdout)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := stdout.String()
	want := "/tmp/svn-test-home|/tmp/svn-test-ca.pem|http://127.0.0.1:3128|explicit|en_US.UTF-8"
	if got != want {
		t.Fatalf("Run() env = %q, want %q", got, want)
	}
}

func TestSSLArgs(t *testing.T) {
	oldAuthorityFile, oldFailures := SSLAuthorityFiles, TrustServerCertFailures
	t.Cleanup(func() {
		SSLAuthorityFiles, TrustServerCertFailures = oldAuthorityFile, oldFailures
	})

	SSLAuthorityFiles = "/etc/ssl/certs/ca-certificates.crt"
	TrustServerCertFailures = "unknown-ca"

	const authorityArg = "servers:global:ssl-authority-files=/etc/ssl/certs/ca-certificates.crt"

	tests := []struct {
		name       string
		executable string
		subcommand string
		want       []string
	}{
		{
			name:       "svn uses trust-server-cert-failures",
			executable: SvnExecutable,
			subcommand: "info",
			want:       []string{"--config-option", authorityArg, "--trust-server-cert-failures=unknown-ca"},
		},
		{
			name:       "svnsync uses source-trust-server-cert-failures",
			executable: SvnsyncExecutable,
			subcommand: "init",
			want:       []string{"--config-option", authorityArg, "--source-trust-server-cert-failures=unknown-ca"},
		},
		{
			name:       "svnadmin is not injected",
			executable: SvnadminExecutable,
			subcommand: "create",
			want:       nil,
		},
		{
			name:       "no subcommand is not injected",
			executable: SvnExecutable,
			subcommand: "",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sslArgs(tt.executable, tt.subcommand); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("sslArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedactArgs(t *testing.T) {
	t.Parallel()

	args := []string{
		"--username", "user",
		"--password", "secret",
		"--source-password=another",
		"--sync-password", "third",
		"--non-interactive", "info", "https://example.com/svn/repo",
	}

	got := redactArgs(args)
	want := []string{
		"--username", "user",
		"--password", "******",
		"--source-password=******",
		"--sync-password", "******",
		"--non-interactive", "info", "https://example.com/svn/repo",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redactArgs() = %v, want %v", got, want)
	}
	if args[3] != "secret" {
		t.Fatalf("redactArgs() mutated the input args: %v", args)
	}
}

func TestEnvKeys(t *testing.T) {
	t.Parallel()

	got := envKeys([]string{"HOME=/root", "HTTPS_PROXY=http://user:pass@proxy", "LANG=en_US.UTF-8"})
	want := []string{"HOME", "HTTPS_PROXY", "LANG"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("envKeys() = %v, want %v", got, want)
	}
}

func TestMakeArgsInjectsSSLArgs(t *testing.T) {
	oldAuthorityFile, oldFailures := SSLAuthorityFiles, TrustServerCertFailures
	t.Cleanup(func() {
		SSLAuthorityFiles, TrustServerCertFailures = oldAuthorityFile, oldFailures
	})

	SSLAuthorityFiles = "/etc/ssl/certs/ca-certificates.crt"
	TrustServerCertFailures = "unknown-ca"

	cmd := New("info", WithArg("https://repo.example.com/svn/repo"))
	args, err := cmd.makeArgs()
	if err != nil {
		t.Fatalf("makeArgs() error = %v", err)
	}

	want := []string{
		"--config-option", "servers:global:ssl-authority-files=/etc/ssl/certs/ca-certificates.crt",
		"--trust-server-cert-failures=unknown-ca",
		"info", "https://repo.example.com/svn/repo",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("makeArgs() = %v, want %v", args, want)
	}
}
