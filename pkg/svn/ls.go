package svn

import (
	"bytes"
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// LsParams contains the parameters for svn ls.
type LsParams struct {
	// Path is the absolute path to a working copy directory.
	Path string

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
	if p.Path == "" {
		return errors.New("path cannot be empty")
	}
	return nil
}

// Ls runs `svn ls` to list directory contents in a working copy.
// It returns the stdout output trimmed of leading/trailing whitespace.
//
// Example:
//
//	xml, err := svc.Ls(ctx, &LsParams{
//	    Path:     "/data/svn/repos/xxx/demo-repo/zentaopms/",
//	    Revision: "130",
//	    XML:      true,
//	})
func (s *Service) Ls(ctx context.Context, params *LsParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	cmd := command.New("ls",
		command.WithArg(params.Path),
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
