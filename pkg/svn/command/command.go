package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

var (
	SvnExecutable      = "svn"
	SvnadminExecutable = "svnadmin"
	SvnsyncExecutable  = "svnsync"
)

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
	if len(c.Envs) > 0 {
		cmd.Env = c.Envs.Args()
	}
	cmd.Env = append(cmd.Env, "LANG=en_US.UTF-8")
	cmd.Env = append(cmd.Env, options.Envs...)
	cmd.Dir = options.Dir
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	logger.GetLogger(ctx).Debugf("svn execute_dir=%s, args: %q, envs: %q", options.Dir, cmd.Args, cmd.Env)
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

func validatePositionalArg(arg string) error {
	if len(arg) > 0 && arg[0] == '-' {
		return fmt.Errorf("positional arg %q cannot start with dash '-': %w", arg, ErrInvalidArg)
	}
	return nil
}
