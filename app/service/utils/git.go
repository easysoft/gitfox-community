package utils

import (
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

// createRPCWriteParams creates base write parameters for git write operations.
// TODO: this function should be in git package and should accept params as interface (contract)
func createRPCWriteParams(
	ctx context.Context,
	urlProvider url.Provider,
	session *auth.Session,
	repo *repo.Repo,
	isInternal bool,
) (git.WriteParams, error) {
	// generate envars (add everything githook CLI needs for execution)
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		urlProvider.GetInternalAPIURL(ctx),
		int64(repo.ID),
		session.Principal.ID,
		false,
		isInternal,
	)
	if err != nil {
		return git.WriteParams{}, fmt.Errorf("failed to generate git hook environment variables: %w", err)
	}

	return git.WriteParams{
		Actor: git.Identity{
			Name:  session.Principal.UID,
			Email: session.Principal.Email,
		},
		RepoUID: repo.GitUID,
		EnvVars: envVars,
	}, nil
}

// CreateRPCExternalWriteParams creates base write parameters for git external write operations.
// External write operations are direct git pushes.
func CreateRPCExternalWriteParams(
	ctx context.Context,
	urlProvider url.Provider,
	session *auth.Session,
	repo *repo.Repo,
) (git.WriteParams, error) {
	return createRPCWriteParams(ctx, urlProvider, session, repo, false)
}

// CreateRPCInternalWriteParams creates base write parameters for git internal write operations.
// Internal write operations are git pushes that originate from the Harness server.
func CreateRPCInternalWriteParams(
	ctx context.Context,
	urlProvider url.Provider,
	session *auth.Session,
	repo *repo.Repo,
) (git.WriteParams, error) {
	return createRPCWriteParams(ctx, urlProvider, session, repo, true)
}

func MapBranch(b git.Branch) (branch.Branch, error) {
	var commit *commit.Commit
	if b.Commit != nil {
		var err error
		commit, err = MapCommit(b.Commit)
		if err != nil {
			return branch.Branch{}, err
		}
	}
	return branch.Branch{
		Name:   b.Name,
		SHA:    b.SHA.String(),
		Commit: commit,
	}, nil
}

func MapCommit(c *git.Commit) (*commit.Commit, error) {
	if c == nil {
		return nil, fmt.Errorf("commit is nil")
	}

	author, err := MapSignature(&c.Author)
	if err != nil {
		return nil, fmt.Errorf("failed to map author: %w", err)
	}

	committer, err := MapSignature(&c.Committer)
	if err != nil {
		return nil, fmt.Errorf("failed to map committer: %w", err)
	}

	parentSHAs := make([]string, len(c.ParentSHAs))
	for i, sha := range c.ParentSHAs {
		parentSHAs[i] = sha.String()
	}

	return &commit.Commit{
			SHA:        c.SHA.String(),
			ParentSHAs: parentSHAs,
			Title:      c.Title,
			Message:    c.Message,
			Author:     *author,
			Committer:  *committer,
			Stats:      mapStats(c),
		},
		nil
}

func mapStats(c *git.Commit) *commit.CommitStats {
	if len(c.FileStats) == 0 {
		return nil
	}

	var insertions int64
	var deletions int64
	for _, stat := range c.FileStats {
		insertions += stat.Insertions
		deletions += stat.Deletions
	}

	return &commit.CommitStats{
		Total: commit.ChangeStats{
			Insertions: insertions,
			Deletions:  deletions,
			Changes:    insertions + deletions,
		},
		// Files: mapFileStats(c),
	}
}

func MapSignature(s *git.Signature) (*commit.Signature, error) {
	return &commit.Signature{
		Identity: commit.Identity{
			Name:  s.Identity.Name,
			Email: s.Identity.Email,
		},
		When: s.When,
	}, nil
}

func MapToRPCSortOrder(o string) git.SortOrder {
	switch o {
	case "asc":
		return git.SortOrderAsc
	case "desc":
		return git.SortOrderDesc
	default:
		return git.SortOrderDefault
	}
}

func MapRenameDetails(c *git.RenameDetails) *commit.RenameDetails {
	if c == nil {
		return nil
	}
	return &commit.RenameDetails{
		OldPath:         c.OldPath,
		NewPath:         c.NewPath,
		CommitShaBefore: c.CommitShaBefore.String(),
		CommitShaAfter:  c.CommitShaAfter.String(),
	}
}

func GenerateCommitTitle(msg string) string {
	msgL := strings.Split(msg, "\n")
	if len(msgL) > 1 {
		return msgL[0]
	}
	return msg
}
