package webhook

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

// handleEventTagCreated handles tag created events
// and triggers tag created webhooks for the source repo.
func (s *Service) handleEventTagCreated(ctx context.Context,
	event *events.Event[*gitevents.TagCreatedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerTagCreated, event.ID, event.Payload.Account, event.Payload.RepoID,
		func(principal string, repo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, repo.GitUID, event.Payload.SHA)
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

			return &ReferencePayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:   string(constant.WebhookTriggerTagCreated),
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
					SHA:        event.Payload.SHA,
					Commit:     &commitInfo,
					HeadCommit: &commitInfo,
				},
				ReferenceUpdateSegment: webhookModel.ReferenceUpdateSegment{
					OldSHA: commitModel.NilSHA,
					Forced: false,
				},
			}, nil
		})
}

// handleEventTagUpdated handles tag updated events
// and triggers tag updated webhooks for the source repo.
func (s *Service) handleEventTagUpdated(ctx context.Context,
	event *events.Event[*gitevents.TagUpdatedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerTagUpdated, event.ID, event.Payload.Account, event.Payload.RepoID,
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
					Trigger:   string(constant.WebhookTriggerTagUpdated),
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

// handleEventTagDeleted handles tag deleted events
// and triggers tag deleted webhooks for the source repo.
func (s *Service) handleEventTagDeleted(ctx context.Context,
	event *events.Event[*gitevents.TagDeletedPayload]) error {
	return s.triggerForEventWithRepo(ctx, constant.WebhookTriggerTagDeleted, event.ID, event.Payload.Account, event.Payload.RepoID,
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
					Trigger:   string(constant.WebhookTriggerTagDeleted),
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
