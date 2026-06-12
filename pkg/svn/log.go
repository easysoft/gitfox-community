package svn

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// LogParams contains the parameters for svn log.
type LogParams struct {
	// RepoPath is the local filesystem path to the SVN repository.
	// It will be converted to a file:// URL internally.
	// Use this to query the entire repository (e.g. with ListAllCommits).
	RepoPath string

	// Path is an optional absolute path to a working copy or a subdirectory within it.
	// When set, svn log queries commits that affected this specific path.
	// The path is used directly without any file:// prefix.
	// At least one of RepoPath or Path must be provided.
	Path string

	// Revision is the revision or revision range to query, e.g. "HEAD", "42", or "0:HEAD".
	// When empty, defaults to HEAD (the latest revision).
	Revision string

	// Verbose enables the -v flag, which includes changed paths in the output.
	Verbose bool

	// XML enables XML format output (--xml flag).
	XML bool

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *LogParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.RepoPath == "" && p.Path == "" {
		return errors.New("either repo path or working copy path must be provided")
	}
	return nil
}

// Log runs `svn log` to query commit history of a repository or working copy path.
// It returns the stdout output trimmed of leading/trailing whitespace.
//
// Typical use cases:
//   - List recent commits of full repo: Log(ctx, &LogParams{RepoPath: "/path/to/repo", Revision: "HEAD", Verbose: true})
//   - Get all commits as XML: Log(ctx, &LogParams{RepoPath: "/path/to/repo", Revision: "0:HEAD", Verbose: true, XML: true})
//   - Query commits for a subdirectory in working copy: Log(ctx, &LogParams{Path: "/data/wc/zentaopms/api", Revision: "100:200", Verbose: true})
func (s *Service) Log(ctx context.Context, params *LogParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	// Path takes precedence: use working copy path directly, otherwise assemble file:// URL.
	target := params.Path
	if target == "" {
		target = fmt.Sprintf("file://%s", params.RepoPath)
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

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrLog, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ListAllCommits returns all commits from revision 0 to HEAD in XML format.
// The repoPath is a local filesystem path; it will be converted to a file:// URL internally.
// Internally it runs: svn log -v -r 0:HEAD --xml file://<repoPath>
//
// Example:
//
//	xml, err := svc.ListAllCommits(ctx, "/data/gitfox-community/svn/repos/xxx/demo")
func (s *Service) ListAllCommits(ctx context.Context, repoPath string) (string, error) {
	return s.Log(ctx, &LogParams{
		RepoPath: repoPath,
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
