// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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

const MigrateRepoGitUID = "empty_gituid_"
