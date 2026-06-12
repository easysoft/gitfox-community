package svn

import (
	"context"
)

// Interface defines the SVN operations used by the application layer.
type Interface interface {
	// Info runs `svn info` to query information about a repository URL or local path.
	Info(ctx context.Context, params *InfoParams) (string, error)

	// InfoRevision is a convenience method that returns the HEAD revision.
	InfoRevision(ctx context.Context, url, username, password string) (string, error)

	// GetLatestRevision returns the latest (HEAD) revision number of the repository at the given path.
	// The repoPath is a local filesystem path; it will be converted to a file:// URL internally.
	// It is equivalent to: svn info file://<repoPath> --show-item revision
	GetLatestRevision(ctx context.Context, repoPath, username, password string) (string, error)

	// InfoXML returns XML-formatted info for the given URL at the specified revision.
	InfoXML(ctx context.Context, url, revision, username, password string) (string, error)

	// GetPathInfo returns XML-formatted info for a working copy path at the given revision.
	// Equivalent to: svn info -r <revision> --xml <path>
	GetPathInfo(ctx context.Context, path, revision, username, password string) (string, error)

	// Checkout creates a working copy from the source repository at the target path.
	Checkout(ctx context.Context, params *CheckoutParams) error

	// CheckoutFromRepo creates a working copy from a local SVN repository.
	CheckoutFromRepo(ctx context.Context, repoPath, targetPath string) error

	// Update brings the working copy up to the latest revision.
	Update(ctx context.Context, params *UpdateParams) error

	// CreateRepository creates a new local SVN repository at the given path.
	CreateRepository(ctx context.Context, path string) error

	// RemoveRepository removes the local SVN repository directory.
	RemoveRepository(path string) error

	// RemoveWorkingCopy removes the working copy directory.
	RemoveWorkingCopy(path string) error

	// ResolveRepositoryPaths assembles repository and working copy paths from an encoded directory name.
	ResolveRepositoryPaths(encodedDir string) RepositoryPaths

	// SyncInit initializes svnsync for the target repository.
	SyncInit(ctx context.Context, params *SyncInitParams) error

	// Sync executes svnsync to synchronize revisions.
	Sync(ctx context.Context, params *SyncParams) error

	// SyncCopyRevProps copies revision properties from source to target.
	SyncCopyRevProps(ctx context.Context, params *SyncParams) error

	// Log runs `svn log` to query commit history of a repository.
	Log(ctx context.Context, params *LogParams) (string, error)

	// ListAllCommits returns all commits from revision 0 to HEAD in XML format.
	// Internally runs: svn log -v -r 0:HEAD --xml file://<repoPath>
	ListAllCommits(ctx context.Context, repoPath string) (string, error)

	// Diff runs `svn diff` to show differences between revisions for a working copy file or directory.
	Diff(ctx context.Context, params *DiffParams) (string, error)

	// Ls runs `svn ls` to list directory contents in a working copy.
	Ls(ctx context.Context, params *LsParams) (string, error)

	// Cat runs `svn cat` to output the content of a file at a given revision.
	Cat(ctx context.Context, params *CatParams) (string, error)

	// Blame runs `svn blame` to show line-by-line author and revision info for a file.
	Blame(ctx context.Context, params *BlameParams) (string, error)
}

// Ensure *Service implements Interface.
var _ Interface = (*Service)(nil)
