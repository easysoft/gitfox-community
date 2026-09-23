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

// CatParams contains the parameters for svn cat.
type CatParams struct {
	// SvnUID identifies the working copy to cat from.
	// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string

	// Subpath is an optional relative file path inside the working copy.
	// svn cat requires a file target (not a directory).
	Subpath string

	// Revision is the revision number to query, e.g. "130".
	// Corresponds to: svn cat -r <Revision> ...
	// When empty, defaults to HEAD.
	Revision string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *CatParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Cat runs `svn cat` to output the content of a file at a given revision.
// It returns the stdout output trimmed of leading/trailing whitespace.
func (s *Service) Cat(ctx context.Context, params *CatParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	target := s.workingCopyPath(params.SvnUID)
	if params.Subpath != "" {
		target = filepath.Join(target, params.Subpath)
	}

	cmd := command.New("cat",
		command.WithArg(target),
		withCatAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrCat, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// withCatAuth adds --username and --password flags when set.
func withCatAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
