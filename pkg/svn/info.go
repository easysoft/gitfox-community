package svn

import (
	"bytes"
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// InfoParams contains the parameters for svn info.
type InfoParams struct {
	// URL is the target repository URL or local path to query info for.
	// e.g. "file:///data/gitfox-community/svn/repos/xxx/demo"
	// or "/data/gitfox-community/svn/repos/xxx/demo-repo/zentaopms/api"
	URL string

	// Subpath is an optional relative subpath inside the URL.
	// When set, the target becomes {URL}/{Subpath}.
	Subpath string

	// Revision is the revision number to query. When 0, the HEAD revision is used.
	// Corresponds to: svn info -r <Revision> ...
	Revision string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string

	// ShowItem is the value for --show-item flag.
	// e.g. "revision", "url", "repos-root-url", etc.
	// When empty, the flag is omitted.
	ShowItem string

	// XML enables XML format output (--xml flag).
	XML bool
}

func (p *InfoParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.URL == "" {
		return errors.New("url cannot be empty")
	}
	return nil
}

// Info runs `svn info` to query information about a repository URL or local path.
// It returns the stdout output trimmed of leading/trailing whitespace.
//
// Typical use cases:
//   - Check if remote repo exists: Info(ctx, &InfoParams{URL: "file:///...", ShowItem: "revision"})
//   - Get path info at a specific revision: Info(ctx, &InfoParams{URL: "/path/to/repo/...", Revision: "130", XML: true})
func (s *Service) Info(ctx context.Context, params *InfoParams) (string, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}

	stdout := &bytes.Buffer{}

	target := params.URL
	if params.Subpath != "" {
		target = strings.TrimRight(target, "/") + "/" + strings.TrimLeft(params.Subpath, "/")
	}

	cmd := command.New("info",
		command.WithArg(target),
		withInfoAuth(params.Username, params.Password),
	)

	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}
	if params.XML {
		cmd.Add(command.WithFlag("--xml"))
	}
	if params.ShowItem != "" {
		cmd.Add(command.WithFlag("--show-item", params.ShowItem))
	}

	if err := cmd.Run(ctx, command.WithStdout(stdout)); err != nil {
		return "", errors.Wrapf(ErrInfo, "%v", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// InfoRevision is a convenience method that returns the HEAD revision of the given URL.
// This is equivalent to calling Info with ShowItem: "revision".
func (s *Service) InfoRevision(ctx context.Context, url, username, password string) (string, error) {
	return s.Info(ctx, &InfoParams{
		URL:      url,
		Username: username,
		Password: password,
		ShowItem: "revision",
	})
}

// GetLatestRevision returns the latest (HEAD) revision number of the repository identified by svnUID.
// Internally runs: svn info file://{svnRoot}/repos/{svnUID} --show-item revision
func (s *Service) GetLatestRevision(ctx context.Context, svnUID, username, password string) (string, error) {
	return s.Info(ctx, &InfoParams{
		URL:      s.FileURL(svnUID),
		Username: username,
		Password: password,
		ShowItem: "revision",
	})
}

// InfoXML returns the XML-formatted info for the given URL at the specified revision.
// This is equivalent to calling Info with Revision and XML: true.
func (s *Service) InfoXML(ctx context.Context, url, revision, username, password string) (string, error) {
	return s.Info(ctx, &InfoParams{
		URL:      url,
		Revision: revision,
		Username: username,
		Password: password,
		XML:      true,
	})
}

// GetPathInfo returns XML-formatted info for the working copy identified by svnUID at the given revision.
// It is equivalent to: svn info -r <revision> --xml <workingCopyPath>
func (s *Service) GetPathInfo(ctx context.Context, svnUID, revision, username, password string) (string, error) {
	return s.Info(ctx, &InfoParams{
		URL:      s.workingCopyPath(svnUID),
		Revision: revision,
		Username: username,
		Password: password,
		XML:      true,
	})
}

// withInfoAuth adds --username and --password flags when set.
func withInfoAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
