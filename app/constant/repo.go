package constant

// BranchArchivedPrefix 归档分支前缀，创建分支、分支类型前缀、默认分支均不可使用此前缀。
const BranchArchivedPrefix = "archived/"

// RepoState defines repo state.
type RepoState int

// RepoState enumeration.
const (
	RepoStateActive RepoState = iota
	RepoStateGitImport
	RepoStateMigrateGitPush
	RepoStateMigrateDataImport
)

// String returns the string representation of the RepoState.
func (state RepoState) String() string {
	switch state {
	case RepoStateActive:
		return "active"
	case RepoStateGitImport:
		return "git-import"
	case RepoStateMigrateGitPush:
		return "migrate-git-push"
	case RepoStateMigrateDataImport:
		return "migrate-data-import"
	default:
		return "undefined"
	}
}
