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

// LsParams contains the parameters for svn ls.
type LsParams struct {
	// SvnUID identifies the working copy to list.
	// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string

	// Subpath is an optional relative subdirectory inside the working copy.
	// When set, svn ls targets {wc}/{Subpath} instead of the wc root.
	Subpath string

	// Revision is the revision number to query, e.g. "130".
	// Corresponds to: svn ls -r <Revision> ...
	// When empty, defaults to HEAD.
	Revision string

	// XML enables XML format output (--xml flag).
	XML bool

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *LsParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Ls runs `svn ls` to list directory contents in a working copy.
// It returns the stdout output trimmed of leading/trailing whitespace.
func (s *Service) Ls(ctx context.Context, params *LsParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	target := s.workingCopyPath(params.SvnUID)
	if params.Subpath != "" {
		target = filepath.Join(target, params.Subpath)
	}

	cmd := s.newCommand("ls",
		command.WithFlag("--non-interactive"),
		command.WithArg(target),
		withLsAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}
	if params.XML {
		cmd.Add(command.WithFlag("--xml"))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrLs, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// withLsAuth adds --username and --password flags when set.
func withLsAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
