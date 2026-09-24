// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
)

// DiffParams contains the parameters for svn diff.
type DiffParams struct {
	// SvnUID identifies the working copy to diff against.
	// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string

	// Subpath is an optional relative file/dir path inside the working copy.
	Subpath string

	// Revision is the revision or revision range to diff, e.g. "130" or "0:130".
	// Corresponds to: svn diff -r <Revision> ...
	Revision string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *DiffParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Diff runs `svn diff` to show differences between revisions for a working copy file or directory.
// It returns the stdout output trimmed of leading/trailing whitespace.
func (s *Service) Diff(ctx context.Context, params *DiffParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	target := s.workingCopyPath(params.SvnUID)
	if params.Subpath != "" {
		target = filepath.Join(target, params.Subpath)
	}

	cmd := command.New("diff",
		command.WithArg(target),
		withDiffAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrDiff, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// withDiffAuth adds --username and --password flags when set.
func withDiffAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
