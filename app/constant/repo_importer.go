package constant

type RepoProviderType string

const (
	RepoProviderTypeGitHub    RepoProviderType = "github"
	RepoProviderTypeGitLab    RepoProviderType = "gitlab"
	RepoProviderTypeBitbucket RepoProviderType = "bitbucket"
	RepoProviderTypeStash     RepoProviderType = "stash"
	RepoProviderTypeGitea     RepoProviderType = "gitea"
	RepoProviderTypeGogs      RepoProviderType = "gogs"
	RepoProviderTypeAzure     RepoProviderType = "azure"
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
	}
}
