package svn

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
	"github.com/easysoft/gitfox-community/pkg/utils/file"
)

// ExportParams contains the parameters for exporting an SVN repository.
type ExportParams struct {
	// SvnUID identifies the repository to export.
	SvnUID string

	// Revision is the revision to export. When empty, SVN exports HEAD.
	Revision string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *ExportParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Export runs `svn export` against the local repository and streams the
// exported tree as a ZIP archive to w. No persistent archive file or URL is
// created.
func (s *Service) Export(ctx context.Context, params *ExportParams, w io.Writer) (streamStarted bool, err error) {
	if err := params.Validate(); err != nil {
		return false, err
	}
	if w == nil {
		return false, errors.New("export writer cannot be nil")
	}

	// MkdirTemp atomically creates a directory with a random suffix, so
	// concurrent exports never share a working directory.
	tempDir, err := os.MkdirTemp("", "gitfox-svn-export-*")
	if err != nil {
		return false, errors.Wrapf(ErrExport, "failed to create temporary directory: %v", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil && err == nil {
			err = errors.Wrapf(ErrExport, "failed to remove temporary directory %s: %v", tempDir, cleanupErr)
		}
	}()

	exportDir := filepath.Join(tempDir, "repository")
	cmd := command.New("export",
		command.WithFlag("--non-interactive", "--no-auth-cache"),
		command.WithArg(s.FileURL(params.SvnUID)),
		command.WithArg(exportDir),
		withExportAuth(params.Username, params.Password),
	)
	if params.Revision != "" {
		cmd.Add(command.WithFlag("-r", params.Revision))
	}
	if err := cmd.Run(ctx); err != nil {
		return false, errors.Wrapf(ErrExport, "svn export failed: %v", err)
	}

	// From here on, WriteZipFromDir may flush bytes to w — any error is a
	// mid-stream failure and must be reported with streamStarted=true.
	if err := file.ZipDir(ctx, exportDir, w); err != nil {
		return true, errors.Wrapf(ErrExport, "failed to write ZIP archive: %v", err)
	}
	return true, nil
}

func withExportAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
