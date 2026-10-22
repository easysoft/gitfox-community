// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package command

import "io"

type CmdOptionFunc func(c *Command)

// WithExecutable overrides the default executable name ("svn").
// Use this for svnadmin, svnsync, etc.
func WithExecutable(name string) CmdOptionFunc {
	return func(c *Command) {
		c.Executable = name
	}
}

// WithFlag set optional flags to pass before positional arguments.
func WithFlag(flags ...string) CmdOptionFunc {
	return func(c *Command) {
		c.Flags = append(c.Flags, flags...)
	}
}

// WithArg add arguments that shall be passed after all flags.
func WithArg(args ...string) CmdOptionFunc {
	return func(c *Command) {
		c.Args = append(c.Args, args...)
	}
}

// WithPostSepArg set arguments that shall be passed as positional arguments after the `--`.
func WithPostSepArg(args ...string) CmdOptionFunc {
	return func(c *Command) {
		c.PostSepArgs = append(c.PostSepArgs, args...)
	}
}

// WithEnv sets environment variable using key value pair
// for example: WithEnv("SVN_TRACE", "true").
func WithEnv(keyValPairs ...string) CmdOptionFunc {
	return func(c *Command) {
		for i := 0; i < len(keyValPairs); i += 2 {
			k, v := keyValPairs[i], keyValPairs[i+1]
			c.Envs[k] = v
		}
	}
}

// WithTLS sets the TLS trust settings injected into svn/svnsync commands.
// authorityFiles is a semicolon-separated list of single-certificate CA files;
// trustServerCertFailures is a comma-separated list of certificate failure types
// to ignore. Only svn and svnsync consume these settings; svnadmin ignores them.
func WithTLS(authorityFiles, trustServerCertFailures string) CmdOptionFunc {
	return func(c *Command) {
		c.SSLAuthorityFiles = authorityFiles
		c.TrustServerCertFailures = trustServerCertFailures
	}
}

// RunOption contains option for running a command.
type RunOption struct {
	// Dir is location of repo.
	Dir string
	// Stdin is the input to the command.
	Stdin io.Reader
	// Stdout is the outputs from the command.
	Stdout io.Writer
	// Stderr is the error output from the command.
	Stderr io.Writer
	// Envs is environments slice containing (final) immutable
	// environment pair "ENV=value"
	Envs []string
}

type RunOptionFunc func(option *RunOption)

// WithDir set directory RunOption.Dir, this is directory
// where svn command should be running.
func WithDir(dir string) RunOptionFunc {
	return func(option *RunOption) {
		option.Dir = dir
	}
}

// WithStdin set RunOption.Stdin reader.
func WithStdin(stdin io.Reader) RunOptionFunc {
	return func(option *RunOption) {
		option.Stdin = stdin
	}
}

// WithStdout set RunOption.Stdout writer.
func WithStdout(stdout io.Writer) RunOptionFunc {
	return func(option *RunOption) {
		option.Stdout = stdout
	}
}

// WithStderr set RunOption.Stderr writer.
func WithStderr(stderr io.Writer) RunOptionFunc {
	return func(option *RunOption) {
		option.Stderr = stderr
	}
}

// WithEnvs sets immutable values as slice, it is always added
// at the end of env slice.
func WithEnvs(envs ...string) RunOptionFunc {
	return func(option *RunOption) {
		option.Envs = append(option.Envs, envs...)
	}
}

// Envs custom key value store for environment variables.
type Envs map[string]string

func (e Envs) Args() []string {
	slice := make([]string, 0, len(e))
	for key, val := range e {
		slice = append(slice, key+"="+val)
	}
	return slice
}
