// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package webhook

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	commitModel "github.com/easysoft/gitfox-community/app/model/commit"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	webhookModel "github.com/easysoft/gitfox-community/app/model/webhook"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	gitenum "github.com/easysoft/gitfox-community/pkg/git/enum"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const MaxWebhookCommitFileStats = 20

// ReferencePayload describes the payload of Reference related webhook triggers.
// Note: Use same payload for all reference operations to make it easier for consumers.
type ReferencePayload struct {
	webhookModel.BaseSegment
	webhookModel.ReferenceSegment
	webhookModel.ReferenceDetailsSegment
	webhookModel.ReferenceUpdateSegment
}

// handleEventBranchCreated handles branch created events
// and triggers branch created webhooks for the source repo.
func (s *Service) handleEventBranchCreated(ctx context.Context,
	event *events.Event[*gitevents.BranchCreatedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerBranchCreated, event.ID, event.Payload.Account, event.Payload.RepoID,
		func(principal string, repo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, repo.GitUID, event.Payload.SHA)
			if err != nil {
				return nil, err
			}
			commits := []webhookModel.CommitInfo{commitInfo}
			repoInfo := webhookModel.RepositoryInfo{
				ID:            int64(repo.ID),
				Identifier:    repo.Identifier,
				Path:          repo.Path,
				GitUID:        repo.GitUID,
				Description:   repo.Desc,
				DefaultBranch: repo.DefaultBranch,
				GitURL:        repo.GitURL,
				GitSSHURL:     repo.GitSSHURL,
			}
			return &ReferencePayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:   string(constant.WebhookTriggerBranchCreated),
					Repo:      repoInfo,
					Principal: principal,
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: event.Payload.Ref,
						Repo: repoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:               event.Payload.SHA,
					Commit:            &commitInfo,
					Commits:           &commits,
					TotalCommitsCount: 1,
					HeadCommit:        &commitInfo,
				},
				ReferenceUpdateSegment: webhookModel.ReferenceUpdateSegment{
					OldSHA: commitModel.NilSHA,
					Forced: false,
				},
			}, nil
		})
}

// handleEventBranchUpdated handles branch updated events
// and triggers branch updated webhooks for the source repo.
func (s *Service) handleEventBranchUpdated(ctx context.Context,
	event *events.Event[*gitevents.BranchUpdatedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerBranchUpdated, event.ID, event.Payload.Account, event.Payload.RepoID,
		func(principal string, repo *repoModel.Repo) (any, error) {
			commitsInfo, totalCommits, err := s.fetchCommitsInfoForEvent(ctx, repo.GitUID, event.Payload.OldSHA, event.Payload.NewSHA)
			if err != nil {
				return nil, err
			}

			repoInfo := webhookModel.RepositoryInfo{
				ID:            int64(repo.ID),
				Identifier:    repo.Identifier,
				Path:          repo.Path,
				GitUID:        repo.GitUID,
				Description:   repo.Desc,
				DefaultBranch: repo.DefaultBranch,
				GitURL:        repo.GitURL,
				GitSSHURL:     repo.GitSSHURL,
			}

			var headCommit *webhookModel.CommitInfo
			if len(commitsInfo) > 0 {
				headCommit = &commitsInfo[0]
			}

			return &ReferencePayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:   string(constant.WebhookTriggerBranchUpdated),
					Repo:      repoInfo,
					Principal: principal,
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: event.Payload.Ref,
						Repo: repoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:               event.Payload.NewSHA,
					Commit:            headCommit,
					HeadCommit:        headCommit,
					Commits:           &commitsInfo,
					TotalCommitsCount: totalCommits,
				},
				ReferenceUpdateSegment: webhookModel.ReferenceUpdateSegment{
					OldSHA: event.Payload.OldSHA,
					Forced: event.Payload.Forced,
				},
			}, nil
		})
}

// handleEventBranchDeleted handles branch deleted events
// and triggers branch deleted webhooks for the source repo.
func (s *Service) handleEventBranchDeleted(ctx context.Context,
	event *events.Event[*gitevents.BranchDeletedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerBranchDeleted, event.ID, event.Payload.Account, event.Payload.RepoID,
		func(principal string, repo *repoModel.Repo) (any, error) {
			repoInfo := webhookModel.RepositoryInfo{
				ID:            int64(repo.ID),
				Identifier:    repo.Identifier,
				Path:          repo.Path,
				GitUID:        repo.GitUID,
				Description:   repo.Desc,
				DefaultBranch: repo.DefaultBranch,
				GitURL:        repo.GitURL,
				GitSSHURL:     repo.GitSSHURL,
			}

			return &ReferencePayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:   string(constant.WebhookTriggerBranchDeleted),
					Repo:      repoInfo,
					Principal: principal,
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: event.Payload.Ref,
						Repo: repoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:        commitModel.NilSHA,
					Commit:     nil,
					HeadCommit: nil,
				},
				ReferenceUpdateSegment: webhookModel.ReferenceUpdateSegment{
					OldSHA: event.Payload.SHA,
					Forced: false,
				},
			}, nil
		})
}

func (s *Service) fetchCommitInfoForEvent(ctx context.Context, repoUID string, commitSHA string) (webhookModel.CommitInfo, error) {
	out, err := s.git.GetCommit(ctx, &git.GetCommitParams{
		ReadParams: git.ReadParams{
			RepoUID: repoUID,
		},
		Revision: commitSHA,
	})

	if err != nil {
		return webhookModel.CommitInfo{}, fmt.Errorf("failed to get commit with targetSha '%s': %w", commitSHA, err)
	}

	return commitInfoFrom(ctx, out.Commit), nil
}

func (s *Service) fetchCommitsInfoForEvent(
	ctx context.Context,
	repoUID string,
	oldSHA string,
	newSHA string,
) ([]webhookModel.CommitInfo, int, error) {
	out, err := s.git.ListCommits(ctx, &git.ListCommitsParams{
		ReadParams:   git.ReadParams{RepoUID: repoUID},
		GitREF:       newSHA,
		After:        oldSHA,
		Page:         1,
		Limit:        MaxWebhookCommitFileStats,
		IncludeStats: true,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list commits from '%s' to '%s': %w", oldSHA, newSHA, err)
	}

	commitsInfo := make([]webhookModel.CommitInfo, len(out.Commits))
	for i := range out.Commits {
		commitsInfo[i] = commitInfoFrom(ctx, out.Commits[i])
	}

	return commitsInfo, out.TotalCommits, nil
}

// commitInfoFrom gets the CommitInfo from a git.Commit.
func commitInfoFrom(ctx context.Context, commit git.Commit) webhookModel.CommitInfo {
	log := logger.GetLogger(ctx)
	added := []string{}
	removed := []string{}
	modified := []string{}

	for _, stat := range commit.FileStats {
		switch {
		case stat.Status == gitenum.FileDiffStatusModified:
			modified = append(modified, stat.Path)
		case stat.Status == gitenum.FileDiffStatusRenamed:
			added = append(added, stat.Path)
			removed = append(removed, stat.OldPath)
		case stat.Status == gitenum.FileDiffStatusDeleted:
			removed = append(removed, stat.Path)
		case stat.Status == gitenum.FileDiffStatusAdded || stat.Status == gitenum.FileDiffStatusCopied:
			added = append(added, stat.Path)
		case stat.Status == gitenum.FileDiffStatusUndefined:
		default:
			log.Warnf("unknown status %q for path %q", stat.Status, stat.Path)
		}
	}

	return webhookModel.CommitInfo{
		SHA:       commit.SHA.String(),
		Message:   commit.Message,
		Author:    signatureInfoFrom(commit.Author),
		Committer: signatureInfoFrom(commit.Committer),
		Added:     added,
		Removed:   removed,
		Modified:  modified,
	}
}

// signatureInfoFrom gets the SignatureInfo from a git.Signature.
func signatureInfoFrom(signature git.Signature) webhookModel.SignatureInfo {
	return webhookModel.SignatureInfo{
		Identity: identityInfoFrom(signature.Identity),
		When:     signature.When,
	}
}

// identityInfoFrom gets the IdentityInfo from a git.Identity.
func identityInfoFrom(identity git.Identity) webhookModel.IdentityInfo {
	return webhookModel.IdentityInfo{
		Name:  identity.Name,
		Email: identity.Email,
	}
}
