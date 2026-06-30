package svn

import (
	"context"
)

// Interface defines the SVN operations used by the application layer.
type Interface interface {
	// FileURL returns the file:// URL for a repository: file://{svnRoot}/repos/{svnUID}
	FileURL(svnUID string) string

	// Info runs `svn info` to query information about a repository URL or local path.
	Info(ctx context.Context, params *InfoParams) (string, error)

	// InfoRevision is a convenience method that returns the HEAD revision.
	InfoRevision(ctx context.Context, url, username, password string) (string, error)

	// GetLatestRevision returns the latest (HEAD) revision number of the repository identified by svnUID.
	// Internally constructs a file:// URL from the svnUID and runs: svn info file://<repoPath> --show-item revision
	GetLatestRevision(ctx context.Context, svnUID, username, password string) (string, error)

	// InfoXML returns XML-formatted info for the given URL at the specified revision.
	InfoXML(ctx context.Context, url, revision, username, password string) (string, error)

	// GetPathInfo returns XML-formatted info for the working copy identified by svnUID at the given revision.
	// Equivalent to: svn info -r <revision> --xml <workingCopyPath>
	GetPathInfo(ctx context.Context, svnUID, revision, username, password string) (string, error)

	// Checkout creates a working copy identified by svnUID from the local SVN repository.
	Checkout(ctx context.Context, params *CheckoutParams) error

	// Update brings the working copy identified by svnUID up to the latest revision.
	Update(ctx context.Context, params *UpdateParams) error

	// CreateRepository creates a new local SVN repository for the given svnUID.
	CreateRepository(ctx context.Context, svnUID string) error

	// RemoveRepository removes the local SVN repository directory for the given svnUID.
	RemoveRepository(svnUID string) error

	// RemoveWorkingCopy removes the working copy directory for the given svnUID.
	RemoveWorkingCopy(svnUID string) error

	// SyncInit initializes svnsync for the target repository identified by svnUID.
	SyncInit(ctx context.Context, params *SyncInitParams) error

	// Sync executes svnsync to synchronize revisions for the repository identified by svnUID.
	Sync(ctx context.Context, params *SyncParams) error

	// SyncCopyRevProps copies revision properties from source to target for the repository identified by svnUID.
	SyncCopyRevProps(ctx context.Context, params *SyncParams) error

	// Log runs `svn log` to query commit history of a repository.
	Log(ctx context.Context, params *LogParams) (string, error)

	// ListAllCommits returns all commits from revision 0 to HEAD in XML format for the repository identified by svnUID.
	ListAllCommits(ctx context.Context, svnUID string) (string, error)

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
