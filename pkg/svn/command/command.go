// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

var (
	SvnExecutable      = "svn"
	SvnadminExecutable = "svnadmin"
	SvnsyncExecutable  = "svnsync"

	// inheritedEnvAllowlist 列出允许从父进程透传给 svn 子进程的环境变量。
	// 这些变量不属于业务敏感数据，但对 svn 走网络/读用户配置非常关键：
	//   - HTTP(S)_PROXY / NO_PROXY / ALL_PROXY：serf 识别的代理配置，
	//     缺失会导致 svn 在需要代理的环境下 TLS 中断。
	//   - SSL_CERT_FILE / SSL_CERT_DIR：自签名/私有 CA 证书路径。
	//   - HOME（Windows 为 USERPROFILE/APPDATA）：svn 需要读 ~/.subversion/servers
	//     与 ~/.subversion/auth 下的证书信任缓存。
	inheritedEnvAllowlist = []string{
		"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY",
		"http_proxy", "https_proxy", "no_proxy", "all_proxy",
		"SSL_CERT_FILE", "SSL_CERT_DIR",
		"HOME", "USERPROFILE", "APPDATA",
		"PATH", "TMP", "TEMP", "TMPDIR", "SSH_AUTH_SOCK",
	}
)

// inheritedEnv 从当前进程按白名单收集环境变量，返回 "KEY=VALUE" 形式的 slice。
func inheritedEnv() []string {
	out := make([]string, 0, len(inheritedEnvAllowlist))
	for _, k := range inheritedEnvAllowlist {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, k+"="+v)
		}
	}
	return out
}

// Command contains options for running an svn command.
type Command struct {
	// Executable is the name of the binary to invoke, e.g. "svn", "svnadmin".
	// Defaults to SvnExecutable ("svn") when empty.
	Executable string

	// Name is the name of the SVN sub-command to run, e.g. "checkout", "update".
	// When empty (e.g. for "svn --version"), no sub-command is inserted.
	Name string

	// Flags is the number of optional flags to pass before positional arguments, e.g.
	// "--version", "--quiet".
	Flags []string

	// Args is the arguments that shall be passed after all flags. These arguments must not be
	// flags and thus cannot start with `-`.
	Args []string

	// PostSepArgs is the arguments that shall be passed as positional arguments after the `--`
	// separator.
	PostSepArgs []string

	// Envs holds environment variables for the command.
	Envs Envs

	// SSLAuthorityFiles 非空时，svn/svnsync 会通过
	// --config-option servers:global:ssl-authority-files=<paths> 显式信任这些 CA 文件。
	// 多个路径以分号分隔；svn 每个文件只读取第一张证书，因此这里应放单证书文件。
	SSLAuthorityFiles string

	// TrustServerCertFailures 非空时，svn 会带上 --trust-server-cert-failures，
	// svnsync 会带上 --source-trust-server-cert-failures，忽略指定类型的证书校验失败。
	// 仅在显式配置时生效，会降低 TLS 安全性。
	TrustServerCertFailures string

	mux sync.RWMutex
}

// New creates new command for interacting with the svn process.
func New(name string, options ...CmdOptionFunc) *Command {
	c := &Command{
		Executable: SvnExecutable,
		Name:       name,
		Envs:       make(Envs),
	}

	c.Add(options...)

	return c
}

// Add appends given options to the command.
func (c *Command) Add(options ...CmdOptionFunc) *Command {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, opt := range options {
		opt(c)
	}
	return c
}

// Run executes the svn command with optional configuration using WithXxx functions.
func (c *Command) Run(ctx context.Context, opts ...RunOptionFunc) (err error) {
	options := &RunOption{}
	for _, f := range opts {
		f(options)
	}

	if options.Stdout == nil {
		options.Stdout = io.Discard
	}
	errAsBuff := false
	if options.Stderr == nil {
		options.Stderr = new(bytes.Buffer)
		errAsBuff = true
	}

	args, err := c.makeArgs()
	if err != nil {
		return fmt.Errorf("failed to build argument list: %w", err)
	}
	cmd := exec.CommandContext(ctx, c.Executable, args...)
	// 从父进程按白名单继承代理/证书/HOME，避免 svn 子进程失去用户级信任配置。
	cmd.Env = inheritedEnv()
	if len(c.Envs) > 0 {
		cmd.Env = append(cmd.Env, c.Envs.Args()...)
	}
	cmd.Env = append(cmd.Env, "LANG=en_US.UTF-8")
	cmd.Env = append(cmd.Env, options.Envs...)
	cmd.Dir = options.Dir
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	logger.GetLogger(ctx).Debugf("svn execute_dir=%s, args: %q, env_keys: %q",
		options.Dir, redactArgs(cmd.Args), envKeys(cmd.Env))
	if err = cmd.Start(); err != nil {
		return err
	}

	result := make(chan error)
	go func() {
		result <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		<-result
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err := cmd.Process.Kill(); err != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("kill process: %w", err)
			}
		}

		return ctx.Err()
	case err = <-result:
		if err == nil {
			return nil
		}

		var stderr []byte
		if buff, ok := options.Stderr.(*bytes.Buffer); ok && errAsBuff {
			stderr = buff.Bytes()
		}
		return NewError(err, stderr)
	}
}

func (c *Command) makeArgs() ([]string, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	var safeArgs []string

	// add flags
	if len(c.Flags) > 0 {
		safeArgs = append(safeArgs, c.Flags...)
	}

	// add configured TLS trust options before the sub-command
	safeArgs = append(safeArgs, c.sslArgs()...)

	// add sub-command name if present
	if c.Name != "" {
		safeArgs = append(safeArgs, c.Name)
	}

	// validate and add positional args
	for _, a := range c.Args {
		if err := validatePositionalArg(a); err != nil {
			return nil, err
		}
	}
	safeArgs = append(safeArgs, c.Args...)

	// add post-separator args
	if len(c.PostSepArgs) > 0 {
		safeArgs = append(safeArgs, "--")
		safeArgs = append(safeArgs, c.PostSepArgs...)
	}

	return safeArgs, nil
}

// sslArgs 返回需要注入 svn/svnsync 的 TLS 信任参数。
// 参数必须放在子命令之前；svnadmin 不支持这些选项，且无子命令（如 --version）时无需注入。
func (c *Command) sslArgs() []string {
	if c.Name == "" {
		return nil
	}

	switch c.Executable {
	case SvnExecutable:
		var args []string
		if c.SSLAuthorityFiles != "" {
			args = append(args, "--config-option", "servers:global:ssl-authority-files="+c.SSLAuthorityFiles)
		}
		if c.TrustServerCertFailures != "" {
			args = append(args, "--trust-server-cert-failures="+c.TrustServerCertFailures)
		}
		return args
	case SvnsyncExecutable:
		var args []string
		if c.SSLAuthorityFiles != "" {
			args = append(args, "--config-option", "servers:global:ssl-authority-files="+c.SSLAuthorityFiles)
		}
		if c.TrustServerCertFailures != "" {
			args = append(args, "--source-trust-server-cert-failures="+c.TrustServerCertFailures)
		}
		return args
	default:
		return nil
	}
}

// sensitiveArgFlags 列出需要在日志中脱敏的参数（密码可能以 "--flag value"
// 或 "--flag=value" 两种形式出现）。
var sensitiveArgFlags = []string{"--password", "--source-password", "--sync-password"}

// redactArgs 返回脱敏后的参数副本，避免密码写入 debug 日志。
func redactArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)

	for i := 0; i < len(out); i++ {
		for _, flag := range sensitiveArgFlags {
			if out[i] == flag && i+1 < len(out) {
				out[i+1] = "******"
				i++
				break
			}
			if strings.HasPrefix(out[i], flag+"=") {
				out[i] = flag + "=******"
				break
			}
		}
	}
	return out
}

// envKeys 只返回环境变量的 key，避免代理 URL 等值中的凭据进入日志。
func envKeys(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if k, _, ok := strings.Cut(kv, "="); ok {
			out = append(out, k)
		} else {
			out = append(out, kv)
		}
	}
	return out
}

func validatePositionalArg(arg string) error {
	if len(arg) > 0 && arg[0] == '-' {
		return fmt.Errorf("positional arg %q cannot start with dash '-': %w", arg, ErrInvalidArg)
	}
	return nil
}
