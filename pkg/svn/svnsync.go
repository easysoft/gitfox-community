package svn

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// SyncInitParams contains the parameters for initializing svnsync.
type SyncInitParams struct {
	// SvnUID identifies the target SVN repository.
	// The target file:// URL is constructed internally as file://{svnRoot}/repos/{svnUID}.
	SvnUID string

	// SourceURL is the remote SVN repository URL to sync from.
	SourceURL string

	// SourceUsername is the username for authenticating to the source repository.
	SourceUsername string

	// SourcePassword is the password for authenticating to the source repository.
	SourcePassword string
}

func (p *SyncInitParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	if p.SourceURL == "" {
		return errors.New("source url cannot be empty")
	}
	return nil
}

// SyncInit initializes svnsync for the target repository so it can sync from the source URL.
// This corresponds to: svnsync init file://{svnRoot}/repos/{svnUID} <SourceURL> --source-username ... --source-password ...
func (s *Service) SyncInit(ctx context.Context, params *SyncInitParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	cmd := command.New("init",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(s.FileURL(params.SvnUID)),
		command.WithArg(params.SourceURL),
		withSyncSourceAuth(params.SourceUsername, params.SourcePassword),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrSyncInit, "%v", err)
	}
	return nil
}

// withSyncSourceAuth adds --source-username and --source-password flags when set.
func withSyncSourceAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--source-username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--source-password", password)
		}
	}
}

// SyncParams contains the parameters for syncing an svnsync-managed repository.
type SyncParams struct {
	// SvnUID identifies the target SVN repository.
	// The target file:// URL is constructed internally as file://{svnRoot}/repos/{svnUID}.
	SvnUID string

	// SourceUsername is the username for authenticating to the source repository.
	SourceUsername string

	// SourcePassword is the password for authenticating to the source repository.
	SourcePassword string
}

func (p *SyncParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Sync executes svnsync sync to synchronize revisions from the source to the target repository.
// This corresponds to: svnsync sync file://{svnRoot}/repos/{svnUID} --source-username ... --source-password ...
func (s *Service) Sync(ctx context.Context, params *SyncParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	cmd := command.New("sync",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(s.FileURL(params.SvnUID)),
		withSyncSourceAuth(params.SourceUsername, params.SourcePassword),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrSync, "%v", err)
	}
	return nil
}

// SyncCopyRevProps copies revision properties (svn:author, svn:log, svn:date, and custom
// revprops) from the source to the target repository. This should be called after Sync to
// ensure the target has complete commit metadata.
// This corresponds to: svnsync copy-revprops file://{svnRoot}/repos/{svnUID} --source-username ... --source-password ...
func (s *Service) SyncCopyRevProps(ctx context.Context, params *SyncParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	cmd := command.New("copy-revprops",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(s.FileURL(params.SvnUID)),
		withSyncSourceAuth(params.SourceUsername, params.SourcePassword),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrSyncCopyRevProps, "%v", err)
	}
	return nil
}
