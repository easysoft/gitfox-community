package svn

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// SyncInitParams contains the parameters for initializing svnsync.
type SyncInitParams struct {
	// TargetRepoPath is the local filesystem path to the target SVN repository.
	// It will be converted to a file:// URL internally.
	TargetRepoPath string

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
	if p.TargetRepoPath == "" {
		return errors.New("target repo path cannot be empty")
	}
	if p.SourceURL == "" {
		return errors.New("source url cannot be empty")
	}
	return nil
}

// SyncInit initializes svnsync for the target repository so it can sync from the source URL.
// This corresponds to: svnsync init file://<TargetRepoPath> <SourceURL> --source-username ... --source-password ...
func (s *Service) SyncInit(ctx context.Context, params *SyncInitParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	targetURL := fmt.Sprintf("file://%s", params.TargetRepoPath)

	cmd := command.New("init",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(targetURL),
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
	// TargetRepoPath is the local filesystem path to the target SVN repository.
	// It will be converted to a file:// URL internally.
	TargetRepoPath string

	// SourceUsername is the username for authenticating to the source repository.
	SourceUsername string

	// SourcePassword is the password for authenticating to the source repository.
	SourcePassword string
}

func (p *SyncParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.TargetRepoPath == "" {
		return errors.New("target repo path cannot be empty")
	}
	return nil
}

// Sync executes svnsync sync to synchronize revisions from the source to the target repository.
// This corresponds to: svnsync sync file://<TargetRepoPath> --source-username ... --source-password ...
func (s *Service) Sync(ctx context.Context, params *SyncParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	targetURL := fmt.Sprintf("file://%s", params.TargetRepoPath)

	cmd := command.New("sync",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(targetURL),
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
// This corresponds to: svnsync copy-revprops file://<TargetRepoPath> --source-username ... --source-password ...
func (s *Service) SyncCopyRevProps(ctx context.Context, params *SyncParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	targetURL := fmt.Sprintf("file://%s", params.TargetRepoPath)

	cmd := command.New("copy-revprops",
		command.WithExecutable(command.SvnsyncExecutable),
		command.WithArg(targetURL),
		withSyncSourceAuth(params.SourceUsername, params.SourcePassword),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrSyncCopyRevProps, "%v", err)
	}
	return nil
}
