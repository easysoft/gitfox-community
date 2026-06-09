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
