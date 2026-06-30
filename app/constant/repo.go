package constant

// BranchArchivedPrefix 归档分支前缀，创建分支、分支类型前缀、默认分支均不可使用此前缀。
const BranchArchivedPrefix = "archived/"

const (
	RepoStateActive     = "active"
	RepoStateArchived   = "archived"
	RepoStateImporting  = "importing"
	RepoStateSyncing    = "syncing"
	RepoStateSyncFailed = "syncFailed"
)

const (
	RepoScmTypeGit = "git"
	RepoScmTypeSvn = "svn"
)
