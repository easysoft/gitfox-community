package svn

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// BlameParams contains the parameters for svn blame.
type BlameParams struct {
	// SvnUID identifies the working copy to blame.
	// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string

	// Subpath is an optional relative subpath inside the working copy.
	// When set, blame targets {wc}/{Subpath}.
	Subpath string

	// Revision is the revision number to query, e.g. "130".
	// Corresponds to: svn blame -r <Revision> ...
	// When empty, defaults to HEAD.
	Revision string

	// XML enables XML format output (--xml flag).
	XML bool

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *BlameParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Blame runs `svn blame` to show line-by-line author and revision info for a file.
// It returns the stdout output trimmed of leading/trailing whitespace.
func (s *Service) Blame(ctx context.Context, params *BlameParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	target := s.workingCopyPath(params.SvnUID)
	if params.Subpath != "" {
		target = filepath.Join(target, params.Subpath)
	}

	cmd := command.New("blame",
		command.WithArg(target),
		withBlameAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}
	if params.XML {
		cmd.Add(command.WithFlag("--xml"))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrBlame, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// withBlameAuth adds --username and --password flags when set.
func withBlameAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
