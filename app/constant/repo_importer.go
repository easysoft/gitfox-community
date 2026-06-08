package constant

type RepoProviderType string

const (
	RepoProviderTypeGitHub    RepoProviderType = "Github"
	RepoProviderTypeGitLab    RepoProviderType = "Gitlab"
	RepoProviderTypeGitLabCOM RepoProviderType = "Gitlab.com"
	RepoProviderTypeGitLabCN  RepoProviderType = "Gitlab.cn"
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
		RepoProviderTypeGitLabCOM,
		RepoProviderTypeGitLabCN,
		RepoProviderTypeBitbucket,
		RepoProviderTypeStash,
		RepoProviderTypeGitea,
		RepoProviderTypeGogs,
		RepoProviderTypeAzure,
		RepoProviderTypeSVN,
	}
}
