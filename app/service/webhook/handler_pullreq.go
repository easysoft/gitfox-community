package webhook

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

const (
	// gitReferenceNamePrefixBranch is the prefix of references of type branch.
	gitReferenceNamePrefixBranch = "refs/heads/"
)

// principalDisplayName 根据 principal(用户账号) 查询用户表 zt_user 的 realname 作为显示名，
// 查询不到或查询失败时回退为 principal 本身，避免影响 webhook 触发。
func (s *Service) principalDisplayName(ctx context.Context, principal string) string {
	user, err := s.userStore.FindByAccount(ctx, principal)
	if err != nil || user == nil || user.DisplayName == "" {
		return principal
	}
	return user.DisplayName
}

// PullReqCreatedPayload describes the body of the pullreq created trigger.
// TODO: move in separate package for small import?
type PullReqCreatedPayload struct {
	webhookModel.BaseSegment
	webhookModel.PullReqSegment
	webhookModel.PullReqTargetReferenceSegment
	webhookModel.ReferenceSegment
	webhookModel.ReferenceDetailsSegment
	Reviewers []string `json:"reviewers"`
}

// handleEventPullReqCreated handles created events for pull requests
// and triggers pullreq created webhooks for the source repo.
func (s *Service) handleEventPullReqCreated(ctx context.Context,
	event *events.Event[*pullreqevents.CreatedPayload]) error {
	return s.triggerForEventWithPullReq(ctx, constant.WebhookTriggerPullReqCreated, event.ID, event.Payload.Account, event.Payload.TargetRepoID, event.Payload.PullReqID,
		func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, sourceRepo.GitUID, event.Payload.SourceSHA)
			if err != nil {
				return nil, err
			}
			// 优先取 ppm 表（ops_ppm）的 reviewers 字段作为审批人，为空时回退查询评审人表（ops_request_reviewers）
			reviewers, err := s.pullreqStore.GetReviewersByRequestID(ctx, uint(pr.ID))
			if err != nil {
				return nil, err
			}
			if len(reviewers) == 0 {
				reviewers, err = s.pullreqStore.ListReviewerAccountsByRequestID(ctx, uint(pr.ID))
				if err != nil {
					return nil, err
				}
			}
			targetRepoInfo := repositoryInfoFrom(targetRepo)
			sourceRepoInfo := repositoryInfoFrom(sourceRepo)
			return &PullReqCreatedPayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:              string(constant.WebhookTriggerPullReqCreated),
					Repo:                 targetRepoInfo,
					Principal:            principal,
					PrincipalDisplayName: s.principalDisplayName(ctx, principal),
				},
				PullReqSegment: webhookModel.PullReqSegment{
					PullReq: pullReqInfoFrom(pr),
				},
				PullReqTargetReferenceSegment: webhookModel.PullReqTargetReferenceSegment{
					TargetRef: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.TargetBranch,
						Repo: targetRepoInfo,
					},
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.SourceBranch,
						Repo: sourceRepoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:        event.Payload.SourceSHA,
					Commit:     &commitInfo,
					HeadCommit: &commitInfo,
				},
				Reviewers: reviewers,
			}, nil
		})
}

// PullReqReopenedPayload describes the body of the pullreq reopened trigger.
// Note: same as payload for created.
type PullReqReopenedPayload PullReqCreatedPayload

// handleEventPullReqReopened handles reopened events for pull requests
// and triggers pullreq reopened webhooks for the source repo.
func (s *Service) handleEventPullReqReopened(ctx context.Context,
	event *events.Event[*pullreqevents.ReopenedPayload]) error {
	return s.triggerForEventWithPullReq(ctx, constant.WebhookTriggerPullReqReopened, event.ID, event.Payload.Account, event.Payload.TargetRepoID, event.Payload.PullReqID,
		func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, sourceRepo.GitUID, event.Payload.SourceSHA)
			if err != nil {
				return nil, err
			}
			targetRepoInfo := repositoryInfoFrom(targetRepo)
			sourceRepoInfo := repositoryInfoFrom(sourceRepo)

			return &PullReqReopenedPayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:              string(constant.WebhookTriggerPullReqReopened),
					Repo:                 targetRepoInfo,
					Principal:            principal,
					PrincipalDisplayName: s.principalDisplayName(ctx, principal),
				},
				PullReqSegment: webhookModel.PullReqSegment{
					PullReq: pullReqInfoFrom(pr),
				},
				PullReqTargetReferenceSegment: webhookModel.PullReqTargetReferenceSegment{
					TargetRef: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.TargetBranch,
						Repo: targetRepoInfo,
					},
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.SourceBranch,
						Repo: sourceRepoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:        event.Payload.SourceSHA,
					Commit:     &commitInfo,
					HeadCommit: &commitInfo,
				},
			}, nil
		})
}

type PullReqBranchUpdatedPayload struct {
	webhookModel.BaseSegment
	webhookModel.PullReqSegment
	webhookModel.PullReqTargetReferenceSegment
	webhookModel.ReferenceSegment
	webhookModel.ReferenceDetailsSegment
	webhookModel.ReferenceUpdateSegment
}

// handleEventPullReqBranchUpdated handles branch updated events for pull requests
// and triggers pullreq branch updated webhooks for the source repo.
func (s *Service) handleEventPullReqBranchUpdated(ctx context.Context,
	event *events.Event[*pullreqevents.BranchUpdatedPayload]) error {
	return s.triggerForEventWithPullReq(ctx, constant.WebhookTriggerPullReqBranchUpdated, event.ID, event.Payload.Account, event.Payload.TargetRepoID, event.Payload.PullReqID,
		func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error) {
			commitsInfo, totalCommits, err := s.fetchCommitsInfoForEvent(ctx, sourceRepo.GitUID, event.Payload.OldSHA, event.Payload.NewSHA)
			if err != nil {
				return nil, err
			}

			targetRepoInfo := repositoryInfoFrom(targetRepo)
			sourceRepoInfo := repositoryInfoFrom(sourceRepo)

			var headCommit *webhookModel.CommitInfo
			if len(commitsInfo) > 0 {
				headCommit = &commitsInfo[0]
			}

			return &PullReqBranchUpdatedPayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:              string(constant.WebhookTriggerPullReqBranchUpdated),
					Repo:                 targetRepoInfo,
					Principal:            principal,
					PrincipalDisplayName: s.principalDisplayName(ctx, principal),
				},
				PullReqSegment: webhookModel.PullReqSegment{
					PullReq: pullReqInfoFrom(pr),
				},
				PullReqTargetReferenceSegment: webhookModel.PullReqTargetReferenceSegment{
					TargetRef: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.TargetBranch,
						Repo: targetRepoInfo,
					},
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.SourceBranch,
						Repo: sourceRepoInfo,
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

// PullReqClosedPayload describes the body of the pullreq closed trigger.
type PullReqClosedPayload struct {
	webhookModel.BaseSegment
	webhookModel.PullReqSegment
	webhookModel.PullReqTargetReferenceSegment
	webhookModel.ReferenceSegment
	webhookModel.ReferenceDetailsSegment
}

// handleEventPullReqClosed handles closed events for pull requests
// and triggers pullreq closed webhooks for the source repo.
func (s *Service) handleEventPullReqClosed(ctx context.Context,
	event *events.Event[*pullreqevents.ClosedPayload]) error {
	return s.triggerForEventWithPullReq(ctx, constant.WebhookTriggerPullReqClosed, event.ID, event.Payload.Account, event.Payload.TargetRepoID, event.Payload.PullReqID,
		func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, sourceRepo.GitUID, event.Payload.SourceSHA)
			if err != nil {
				return nil, err
			}
			targetRepoInfo := repositoryInfoFrom(targetRepo)
			sourceRepoInfo := repositoryInfoFrom(sourceRepo)

			return &PullReqClosedPayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:              string(constant.WebhookTriggerPullReqClosed),
					Repo:                 targetRepoInfo,
					Principal:            principal,
					PrincipalDisplayName: s.principalDisplayName(ctx, principal),
				},
				PullReqSegment: webhookModel.PullReqSegment{
					PullReq: pullReqInfoFrom(pr),
				},
				PullReqTargetReferenceSegment: webhookModel.PullReqTargetReferenceSegment{
					TargetRef: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.TargetBranch,
						Repo: targetRepoInfo,
					},
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.SourceBranch,
						Repo: sourceRepoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:        event.Payload.SourceSHA,
					Commit:     &commitInfo,
					HeadCommit: &commitInfo,
				},
			}, nil
		})
}

// PullReqMergedPayload describes the body of the pullreq merged trigger.
type PullReqMergedPayload struct {
	webhookModel.BaseSegment
	webhookModel.PullReqSegment
	webhookModel.PullReqTargetReferenceSegment
	webhookModel.ReferenceSegment
	webhookModel.ReferenceDetailsSegment
}

// handleEventPullReqMerged handles merged events for pull requests
// and triggers pullreq merged webhooks for the source repo.
func (s *Service) handleEventPullReqMerged(ctx context.Context,
	event *events.Event[*pullreqevents.MergedPayload]) error {
	return s.triggerForEventWithPullReq(ctx, constant.WebhookTriggerPullReqMerged, event.ID, event.Payload.Account, event.Payload.TargetRepoID, event.Payload.PullReqID,
		func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error) {
			commitInfo, err := s.fetchCommitInfoForEvent(ctx, sourceRepo.GitUID, event.Payload.SourceSHA)
			if err != nil {
				return nil, err
			}
			targetRepoInfo := repositoryInfoFrom(targetRepo)
			sourceRepoInfo := repositoryInfoFrom(sourceRepo)

			return &PullReqMergedPayload{
				BaseSegment: webhookModel.BaseSegment{
					Trigger:              string(constant.WebhookTriggerPullReqMerged),
					Repo:                 targetRepoInfo,
					Principal:            principal,
					PrincipalDisplayName: s.principalDisplayName(ctx, principal),
				},
				PullReqSegment: webhookModel.PullReqSegment{
					PullReq: pullReqInfoFrom(pr),
				},
				PullReqTargetReferenceSegment: webhookModel.PullReqTargetReferenceSegment{
					TargetRef: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.TargetBranch,
						Repo: targetRepoInfo,
					},
				},
				ReferenceSegment: webhookModel.ReferenceSegment{
					Ref: webhookModel.ReferenceInfo{
						Name: gitReferenceNamePrefixBranch + pr.SourceBranch,
						Repo: sourceRepoInfo,
					},
				},
				ReferenceDetailsSegment: webhookModel.ReferenceDetailsSegment{
					SHA:        event.Payload.SourceSHA,
					Commit:     &commitInfo,
					HeadCommit: &commitInfo,
				},
			}, nil
		})
}

func repositoryInfoFrom(repo *repoModel.Repo) webhookModel.RepositoryInfo {
	return webhookModel.RepositoryInfo{
		ID:            int64(repo.ID),
		Path:          repo.Path,
		Identifier:    repo.Identifier,
		Description:   repo.Desc,
		DefaultBranch: repo.DefaultBranch,
		GitURL:        repo.GitURL,
		GitSSHURL:     repo.GitSSHURL,
	}
}

func pullReqInfoFrom(pr *repoModel.PullRequest) webhookModel.PullReqInfo {
	mergeMethod := constant.MergeMethod(pr.MergeMethod)
	return webhookModel.PullReqInfo{
		Number:        int64(pr.ID),
		State:         constant.PullReqState(pr.Status),
		Title:         pr.Title,
		Description:   pr.Desc,
		SourceRepoID:  int64(pr.SourceRepoID),
		SourceBranch:  pr.SourceBranch,
		TargetRepoID:  int64(pr.TargetRepoID),
		TargetBranch:  pr.TargetBranch,
		MergeStrategy: &mergeMethod,
		Author:        pr.CreatedBy,
	}
}
