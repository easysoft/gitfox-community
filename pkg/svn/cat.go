package svn

import (
	"bytes"
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// CatParams contains the parameters for svn cat.
type CatParams struct {
	// Path is the absolute path to a working copy file.
	Path string

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
	if p.Path == "" {
		return errors.New("path cannot be empty")
	}
	return nil
}

// Cat runs `svn cat` to output the content of a file at a given revision.
// It returns the stdout output trimmed of leading/trailing whitespace.
//
// Example:
//
//	content, err := svc.Cat(ctx, &CatParams{
//	    Path:     "/data/svn/repos/xxx/demo-repo/zentaopms/api/v1/lang/zh-cn.php",
//	    Revision: "130",
//	})
func (s *Service) Cat(ctx context.Context, params *CatParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	cmd := command.New("cat",
		command.WithArg(params.Path),
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
