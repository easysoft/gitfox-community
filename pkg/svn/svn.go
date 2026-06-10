package svn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
	"github.com/hashicorp/go-version"
)

// Service provides high-level SVN operations.
type Service struct {
	// traceSvn enables debug logging of svn command execution.
	traceSvn bool
	// svnRoot is the root directory for SVN repository storage.
	svnRoot string
}

// New creates a new SVN Service. It verifies that the svn binary is available on PATH.
func New(cfg *config.Config) (*Service, error) {
	return &Service{svnRoot: cfg.SVN.Root}, nil
}

// Version returns the parsed SVN version from running `svn --version`.
// The raw output looks like:
//
//	svn, version 1.14.2 (r1899510)
//	   compiled Feb 13 2023, 22:39:16 on x86_64-unknown-linux-gnu
func (s *Service) Version(ctx context.Context) (*version.Version, error) {
	cmd := command.New("", command.WithFlag("--version"))
	output := &bytes.Buffer{}
	if err := cmd.Run(ctx, command.WithStdout(output)); err != nil {
		return nil, errors.Wrapf(ErrFailedToGetVersion, "%v", err)
	}
	return parseVersion(output.String())
}

// CheckVersionAtLeast verifies that the installed svn binary version is at least the given
// minimum version (e.g. "1.14.0"). It returns an error if the binary is not found or the
// version is too old.
func (s *Service) CheckVersionAtLeast(ctx context.Context, atLeast string) error {
	svnVersion, err := s.Version(ctx)
	if err != nil {
		return err
	}

	atLeastVersion, err := version.NewVersion(atLeast)
	if err != nil {
		return errors.Newf("invalid required version %q: %v", atLeast, err)
	}

	if svnVersion.Compare(atLeastVersion) < 0 {
		return errors.Wrapf(ErrSvnVersionNotSupported,
			"installed svn version %s is not at least %s", svnVersion.Original(), atLeast)
	}
	return nil
}

// CreateRepository creates a new local SVN repository at the given path using svnadmin create.
// The parent directory will be created if it does not exist.
// After creation, a permissive pre-revprop-change hook is installed so that svnsync can operate.
func (s *Service) CreateRepository(ctx context.Context, path string) error {
	// Ensure parent directory exists.
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0o700); err != nil {
		return errors.Wrapf(ErrCreateRepository, "failed to create parent directory %s: %v", parentDir, err)
	}

	cmd := command.New("create",
		command.WithExecutable(command.SvnadminExecutable),
		command.WithArg(path),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrCreateRepository, "%v", err)
	}

	// svnsync requires the pre-revprop-change hook to be enabled on the target repository.
	if err := s.CreatePreRevpropChangeHook(path); err != nil {
		return errors.Wrapf(ErrCreateRepository, "%v", err)
	}

	return nil
}

// CreatePreRevpropChangeHook creates a permissive pre-revprop-change hook in the repository's
// hooks directory. This hook allows all revision property changes, which is required for
// svnsync to function properly.
func (s *Service) CreatePreRevpropChangeHook(repoPath string) error {
	hookName := "pre-revprop-change"
	var content string

	if runtime.GOOS == "windows" {
		hookName += ".bat"
		content = "@exit /b 0\r\n"
	} else {
		content = "#!/bin/sh\nexit 0\n"
	}

	hookPath := filepath.Join(repoPath, "hooks", hookName)
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		return fmt.Errorf("failed to write hook script: %w", err)
	}

	return nil
}

// RemoveRepository removes the local SVN repository directory at the given path.
// Use this to clean up when import fails at SyncInit, Sync, or Checkout stage.
func (s *Service) RemoveRepository(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.Wrapf(ErrRemoveRepository, "failed to remove repository %s: %v", path, err)
	}
	return nil
}

// RemoveWorkingCopy removes the working copy directory at the given path.
func (s *Service) RemoveWorkingCopy(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.Wrapf(ErrRemoveWorkingCopy, "failed to remove working copy %s: %v", path, err)
	}
	return nil
}

// parseVersion extracts the SVN version from the `svn --version` output.
// Input example:
//
//	svn, version 1.14.2 (r1899510)
//	   compiled ...
//
// It returns the parsed version or an error if parsing fails.
func parseVersion(output string) (*version.Version, error) {
	// First line: "svn, version 1.14.2 (r1899510)"
	// The version string is the 3rd whitespace-separated field.
	firstLine, _, _ := strings.Cut(output, "\n")
	fields := strings.Fields(firstLine)
	if len(fields) < 3 {
		return nil, errors.Newf("invalid svn version output: %s", firstLine)
	}

	// fields[0] = "svn,"
	// fields[1] = "version"
	// fields[2] = "1.14.2" or "1.14.2 (r1899510)" or similar
	versionStr := strings.TrimRight(fields[2], "()")

	parsed, err := version.NewVersion(versionStr)
	if err != nil {
		return nil, errors.Newf("failed to parse svn version %q: %v", versionStr, err)
	}
	return parsed, nil
}
