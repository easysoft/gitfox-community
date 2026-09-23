// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
)

// LogParams contains the parameters for svn log.
type LogParams struct {
	// SvnUID identifies the SVN repository to query.
	// It will be converted to a file:// URL internally.
	// Use this to query the entire repository (e.g. with ListAllCommits).
	SvnUID string

	// Subpath is an optional relative subdirectory inside the repository.
	// When set, the target becomes {file://repo}/{Subpath} so that svn log
	// filters commits touching that path.
	Subpath string

	// Path is an optional absolute path to a working copy or a subdirectory within it.
	// When set, svn log queries commits that affected this specific path.
	// The path is used directly without any file:// prefix.
	// At least one of SvnUID or Path must be provided.
	Path string

	// Revision is the revision or revision range to query, e.g. "HEAD", "42", or "0:HEAD".
	// When empty, defaults to HEAD (the latest revision).
	Revision string

	// Verbose enables the -v flag, which includes changed paths in the output.
	Verbose bool

	// XML enables XML format output (--xml flag).
	XML bool

	// Limit caps the number of log entries returned (`--limit N`).
	// Zero means no limit.
	Limit int

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *LogParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" && p.Path == "" {
		return errors.New("either svnUID or working copy path must be provided")
	}
	return nil
}

// Log runs `svn log` to query commit history of a repository or working copy path.
// It returns the stdout output trimmed of leading/trailing whitespace.
func (s *Service) Log(ctx context.Context, params *LogParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	// Path takes precedence: use working copy path directly, otherwise assemble file:// URL.
	target := params.Path
	if target == "" {
		target = s.FileURL(params.SvnUID)
		if params.Subpath != "" {
			target = strings.TrimRight(target, "/") + "/" + strings.TrimLeft(params.Subpath, "/")
		}
	}

	cmd := command.New("log",
		command.WithArg(target),
		withLogAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}
	if params.Verbose {
		cmd.Add(command.WithFlag("-v"))
	}
	if params.XML {
		cmd.Add(command.WithFlag("--xml"))
	}
	if params.Limit > 0 {
		cmd.Add(command.WithFlag("--limit", strconv.Itoa(params.Limit)))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrLog, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ListAllCommits returns all commits from revision 0 to HEAD in XML format for the repository identified by svnUID.
// Internally runs: svn log -v -r 0:HEAD --xml file://{svnRoot}/repos/{svnUID}
func (s *Service) ListAllCommits(ctx context.Context, svnUID string) (string, error) {
	return s.Log(ctx, &LogParams{
		SvnUID:   svnUID,
		Revision: "0:HEAD",
		Verbose:  true,
		XML:      true,
	})
}

// withLogAuth adds --username and --password flags when set.
func withLogAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
