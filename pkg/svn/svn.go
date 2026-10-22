// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
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
	// caDir is the subdirectory under svnRoot holding split CA certificates.
	caDir = "ca"
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

// New creates a new SVN Service. It applies the optional TLS trust settings from
// config and prepares the configured CA bundle for svn/svnsync.
func New(cfg *config.Config) (*Service, error) {
	command.SSLAuthorityFiles = ""
	command.TrustServerCertFailures = ""

	// 显式 CA 文件与全局 TLS 降级不能同时使用，避免安全兜底掩盖真实问题。
	if cfg.SVN.SSLAuthorityFile != "" && cfg.SVN.TrustServerCertFailures != "" {
		return nil, errors.New("svn ssl authority file and trust-server-cert-failures cannot be configured together")
	}

	if cfg.SVN.TrustServerCertFailures != "" {
		failures, err := normalizeTrustServerCertFailures(cfg.SVN.TrustServerCertFailures)
		if err != nil {
			return nil, err
		}
		command.TrustServerCertFailures = failures
		logger.Warnf("SVN TLS certificate verification is relaxed (trust-server-cert-failures=%s): "+
			"all SVN HTTPS connections will accept certificates with these failures; "+
			"use only for temporary troubleshooting", failures)
	}

	if cfg.SVN.SSLAuthorityFile != "" {
		authorityFiles, err := splitAuthorityFile(cfg.SVN.SSLAuthorityFile, filepath.Join(cfg.SVN.Root, caDir))
		if err != nil {
			return nil, err
		}
		command.SSLAuthorityFiles = authorityFiles
	}

	return &Service{svnRoot: cfg.SVN.Root}, nil
}

var allowedTrustServerCertFailures = map[string]struct{}{
	"unknown-ca":    {},
	"cn-mismatch":   {},
	"expired":       {},
	"not-yet-valid": {},
	"other":         {},
}

// normalizeTrustServerCertFailures 校验并规范化 svn 的 trust-server-cert-failures 配置。
func normalizeTrustServerCertFailures(raw string) (string, error) {
	var values []string
	seen := make(map[string]struct{})
	for part := range strings.SplitSeq(raw, ",") {
		v := strings.ToLower(strings.TrimSpace(part))
		if v == "" {
			continue
		}
		if _, ok := allowedTrustServerCertFailures[v]; !ok {
			return "", errors.Newf(
				"invalid svn trust-server-cert-failures value %q (allowed: unknown-ca, cn-mismatch, expired, not-yet-valid, other)", v)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		values = append(values, v)
	}
	if len(values) == 0 {
		return "", errors.New("svn trust-server-cert-failures must contain at least one value")
	}
	return strings.Join(values, ","), nil
}

// splitAuthorityFile 把 PEM bundle 拆成单证书文件，返回以分号分隔的文件列表。
// svn 的 ssl-authority-files 每个文件只读取第一张证书，直接把系统大 bundle
// 传给它会导致只有第一张证书生效，因此这里先拆分。
func splitAuthorityFile(src, dstDir string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read svn ssl authority file %q", src)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		logger.Warnf("SVN ssl authority file %q is writable by group/others and could be replaced to inject a CA", src)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read svn ssl authority file %q", src)
	}

	// 目录必须存在且不是符号链接，权限收紧到 0700，避免拆分文件被他人替换。
	if info, err := os.Lstat(dstDir); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.Newf("svn ssl authority dir %q must not be a symlink", dstDir)
	}
	if err := os.MkdirAll(dstDir, 0o700); err != nil {
		return "", errors.Wrapf(err, "failed to create svn ssl authority dir %q", dstDir)
	}
	if err := os.Chmod(dstDir, 0o700); err != nil {
		return "", errors.Wrapf(err, "failed to secure svn ssl authority dir %q", dstDir)
	}

	var paths []string
	seen := make(map[string]struct{})
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}

		sum := sha256.Sum256(block.Bytes)
		path := filepath.Join(dstDir, hex.EncodeToString(sum[:16])+".pem")
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}

		if err := writeFileAtomic(path, pem.EncodeToMemory(block), 0o600); err != nil {
			return "", errors.Wrapf(err, "failed to write svn ssl authority file %q", path)
		}
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		return "", errors.Newf("no certificate found in svn ssl authority file %q", src)
	}
	return strings.Join(paths, ";"), nil
}

// writeFileAtomic 通过临时文件 + rename 写入，避免跟随目标路径上已存在的符号链接。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-ca-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
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
