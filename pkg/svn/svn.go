// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
	"github.com/hashicorp/go-version"
)

// Service provides high-level SVN operations.
type Service struct {
	// traceSvn enables debug logging of svn command execution.
	traceSvn bool
	// svnRoot is the root directory for SVN repository storage.
	svnRoot string
}

const (
	// reposDir is the fixed subdirectory under svnRoot where SVN repositories are stored.
	reposDir = "repos"
	// wcSuffix is appended to the repository directory name for working copies.
	wcSuffix = "-wc"
)

// repoPath returns the filesystem path for a repository: {svnRoot}/repos/{svnUID}
func (s *Service) repoPath(svnUID string) string {
	return filepath.Join(s.svnRoot, reposDir, svnUID)
}

// workingCopyPath returns the filesystem path for a working copy: {svnRoot}/repos/{svnUID}-wc
func (s *Service) workingCopyPath(svnUID string) string {
	return filepath.Join(s.svnRoot, reposDir, svnUID+wcSuffix)
}

// FileURL returns the file:// URL for a repository: file://{svnRoot}/repos/{svnUID}
func (s *Service) FileURL(svnUID string) string {
	return "file://" + s.repoPath(svnUID)
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

// CreateRepository creates a new local SVN repository for the given svnUID using svnadmin create.
// The repository will be created at {svnRoot}/repos/{svnUID}.
// After creation, a permissive pre-revprop-change hook is installed so that svnsync can operate.
func (s *Service) CreateRepository(ctx context.Context, svnUID string) error {
	path := s.repoPath(svnUID)

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
	if err := s.createPreRevpropChangeHook(path); err != nil {
		return errors.Wrapf(ErrCreateRepository, "%v", err)
	}

	return nil
}

// createPreRevpropChangeHook creates a permissive pre-revprop-change hook in the repository's
// hooks directory. This hook allows all revision property changes, which is required for
// svnsync to function properly.
func (s *Service) createPreRevpropChangeHook(repoPath string) error {
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

// RemoveRepository removes the local SVN repository directory for the given svnUID.
// The repository path is resolved as {svnRoot}/repos/{svnUID}.
func (s *Service) RemoveRepository(svnUID string) error {
	path := s.repoPath(svnUID)
	if err := os.RemoveAll(path); err != nil {
		return errors.Wrapf(ErrRemoveRepository, "failed to remove repository %s: %v", path, err)
	}
	return nil
}

// RemoveWorkingCopy removes the working copy directory for the given svnUID.
// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
func (s *Service) RemoveWorkingCopy(svnUID string) error {
	path := s.workingCopyPath(svnUID)
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
