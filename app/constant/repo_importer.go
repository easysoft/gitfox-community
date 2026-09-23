// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

type RepoProviderType string

const (
	RepoProviderTypeGitHub    RepoProviderType = "GitHub"
	RepoProviderTypeGitLab    RepoProviderType = "GitLab"
	RepoProviderTypeBitbucket RepoProviderType = "Bitbucket"
	RepoProviderTypeStash     RepoProviderType = "Stash"
	RepoProviderTypeGitea     RepoProviderType = "Gitea"
	RepoProviderTypeGogs      RepoProviderType = "Gogs"
	RepoProviderTypeAzure     RepoProviderType = "Azure"
	RepoProviderTypeSVN       RepoProviderType = "Subversion"
)

func (p RepoProviderType) Enum() []any {
	return []any{
		RepoProviderTypeGitHub,
		RepoProviderTypeGitLab,
		RepoProviderTypeBitbucket,
		RepoProviderTypeStash,
		RepoProviderTypeGitea,
		RepoProviderTypeGogs,
		RepoProviderTypeAzure,
		RepoProviderTypeSVN,
	}
}
