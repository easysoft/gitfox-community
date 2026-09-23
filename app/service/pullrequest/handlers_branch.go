// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pullrequest

import (
	"context"
	"strings"

	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

var (
	errPRNotOpen = errors.New("PR is not open")
)

// triggerPREventOnBranchUpdate handles branch update events. For every open pull request
// it writes an activity entry and triggers the pull request Branch Updated event.
//
//nolint:gocognit // refactor if needed
func (s *Service) triggerPREventOnBranchUpdate(ctx context.Context,
	event *events.Event[*gitevents.BranchUpdatedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming git branch updated event, repoID=%d, ref=%s", event.Payload.RepoID, event.Payload.Ref)

	// we should always update PR mergeable status check when target branch is updated.
	// - main
	//    |- develop
	//         |- feature1
	//         |- feature2
	// when feature2 merge changes into develop branch then feature1 branch is not consistent anymore
	// and need to run mergeable check even nothing was changed on feature1, same applies to main if someone
	// push new commit to main then develop should merge status should be unchecked.
	if branch, err := getBranchFromRef(event.Payload.Ref); err == nil {
		err = s.pullRequestStore.ResetMergeCheckStatus(ctx, uint(event.Payload.RepoID), branch)
		if err != nil {
			log.Errorf("failed to reset merge check repo %v (branch: %v ref:%v) status: %v", event.Payload.RepoID, branch, event.Payload.Ref, err)
			return err
		}
	}
	log.Debugf("reset merge check status for repo %v (ref: %v)", event.Payload.RepoID, event.Payload.Ref)

	var commitTitle string
	err := func() error {
		repo, err := s.repoStore.GetByID(ctx, uint(event.Payload.RepoID))
		if err != nil {
			return errors.Wrapf(err, "failed to get repo info")
		}

		commit, err := s.gitSvc.GetCommit(ctx, &git.GetCommitParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			Revision:   event.Payload.NewSHA,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to get commit info")
		}

		commitTitle = commit.Commit.Title

		return nil
	}()
	if err != nil {
		// non critical error
		log.Warnf("failed to get commit info from git: %v", err)
	}

	// Avoid unused variable warning
	_ = commitTitle

	// TODO: This function is currently executed directly on branch update event.
	// TODO: But it should be executed after the PR's head ref has been updated.
	// TODO: This is to make sure the commit exists on the target repository for forked repositories.
	s.forEveryOpenPR(ctx, event.Payload.RepoID, event.Payload.Ref, func(pr *repoModel.PullRequest) error {
		// First check if the merge base has changed

		targetRepo, err := s.repoStore.GetByID(ctx, pr.TargetRepoID)
		if err != nil {
			return errors.Wrapf(err, "failed to get target repo info")
		}

		mergeBaseInfo, err := s.gitSvc.MergeBase(ctx, git.MergeBaseParams{
			ReadParams: git.ReadParams{RepoUID: targetRepo.GitUID},
			Ref1:       event.Payload.NewSHA,
			Ref2:       pr.TargetBranch,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to get merge base after branch update to=%s for PR=%d",
				event.Payload.NewSHA, pr.ID)
		}

		oldMergeBase := pr.MergeBaseSHA
		newMergeBase := mergeBaseInfo.MergeBaseSHA

		// Update the database with the latest source commit SHA and the merge base SHA.
		pr, err = s.pullRequestStore.UpdateOptLock(ctx, pr, func(pr *repoModel.PullRequest) error {
			// to avoid racing conditions with merge
			if pr.Status != repoModel.PullRequestStatusOpen {
				return errPRNotOpen
			}

			if pr.SourceSHA != event.Payload.OldSHA {
				log.Warnf("failed to set SourceSHA for PR %d to value '%s', expected SHA '%s' but current pr has '%s'",
					pr.ID, event.Payload.NewSHA, event.Payload.OldSHA, pr.SourceSHA)
			}

			pr.SourceSHA = event.Payload.NewSHA
			pr.MergeBaseSHA = newMergeBase.String()

			// reset merge-check fields for new run
			pr.MergeSHA = ""
			pr.CommitCount = 0
			pr.FileCount = 0
			pr.MarkAsMergeUnchecked()

			return nil
		})
		if errors.Is(err, errPRNotOpen) {
			return nil
		}
		if err != nil {
			return err
		}

		// Log the branch update event
		log.Infof("PR %d branch updated: old=%s, new=%s, oldMergeBase=%s, newMergeBase=%s",
			pr.ID, event.Payload.OldSHA, event.Payload.NewSHA, oldMergeBase, newMergeBase.String())

		// Report pull request branch updated event
		s.pullreqEvReporter.BranchUpdated(ctx, &pullreqevents.BranchUpdatedPayload{
			Base: pullreqevents.Base{
				PullReqID:    int64(pr.ID),
				SourceRepoID: int64(pr.SourceRepoID),
				TargetRepoID: int64(pr.TargetRepoID),
				Account:      event.Payload.Account,
			},
			SourceBranch:    pr.SourceBranch,
			OldSHA:          event.Payload.OldSHA,
			NewSHA:          event.Payload.NewSHA,
			OldMergeBaseSHA: oldMergeBase,
			NewMergeBaseSHA: newMergeBase.String(),
			Forced:          event.Payload.Forced,
		})

		// Publish SSE event for real-time UI updates
		if err = s.sseStreamer.Publish(ctx, int64(targetRepo.SpaceID), eventModel.SSETypePullRequestUpdated, pr); err != nil {
			log.Warnf("failed to publish PR changed event: %v", err)
		}

		return nil
	})
	log.Infof("[event]: successfully consumed git branch updated event, repoID=%d, ref=%s", event.Payload.RepoID, event.Payload.Ref)
	return nil
}

// closePullReqOnBranchDelete handles branch delete events.
// It closes every open pull request for the branch and triggers the pull request BranchDeleted event.
func (s *Service) closePullReqOnBranchDelete(ctx context.Context,
	event *events.Event[*gitevents.BranchDeletedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming git branch deleted event, repoID=%d, ref=%s", event.Payload.RepoID, event.Payload.Ref)

	s.forEveryOpenPR(ctx, event.Payload.RepoID, event.Payload.Ref, func(pr *repoModel.PullRequest) error {
		targetRepo, err := s.repoStore.GetByID(ctx, pr.TargetRepoID)
		if err != nil {
			return errors.Wrapf(err, "failed to get repo info")
		}

		pr, err = s.pullRequestStore.UpdateOptLock(ctx, pr, func(pr *repoModel.PullRequest) error {
			// to avoid racing conditions with merge
			if pr.Status != repoModel.PullRequestStatusOpen {
				return errPRNotOpen
			}

			pr.Status = repoModel.PullRequestStatusClosed
			pr.MergeSHA = ""
			pr.MarkAsMergeUnchecked()

			return nil
		})
		if errors.Is(err, errPRNotOpen) {
			return nil
		}
		if err != nil {
			return errors.Wrapf(err, "failed to close pull request after branch delete")
		}

		log.Infof("PR %d closed due to branch deletion", pr.ID)

		// Report pull request closed event
		s.pullreqEvReporter.Closed(ctx, &pullreqevents.ClosedPayload{
			Base: pullreqevents.Base{
				PullReqID:    int64(pr.ID),
				SourceRepoID: int64(pr.SourceRepoID),
				TargetRepoID: int64(pr.TargetRepoID),
				Account:      event.Payload.Account,
			},
			SourceBranch: pr.SourceBranch,
			SourceSHA:    pr.SourceSHA,
		})

		// Publish SSE event for real-time UI updates
		if err = s.sseStreamer.Publish(ctx, int64(targetRepo.SpaceID), eventModel.SSETypePullRequestUpdated, pr); err != nil {
			log.Warnf("failed to publish PR changed event: %v", err)
		}

		return nil
	})
	log.Infof("[event]: successfully consumed git branch deleted event, repoID=%d, ref=%s", event.Payload.RepoID, event.Payload.Ref)
	return nil
}

// forEveryOpenPR is utility function that executes the provided function
// for every open pull request created with the source branch given as a git ref.
func (s *Service) forEveryOpenPR(ctx context.Context,
	repoID int64, ref string,
	fn func(pr *repoModel.PullRequest) error,
) {
	log := logger.GetLogger(ctx)

	branch, err := getBranchFromRef(ref)
	if len(branch) == 0 {
		log.Errorf("failed to get branch from ref: %v", err)
		return
	}
	log.Debugf("forEveryOpenPR: repoID=%d, branch=%s ref=%s", repoID, branch, ref)

	pullreqList, err := s.pullRequestStore.List(ctx, &repoModel.PullRequestFilter{
		SourceRepoID: uint(repoID),
		SourceBranch: branch,
		Status:       repoModel.PullRequestStatusOpen,
	})
	if err != nil {
		log.Errorf("failed to get list of open pull requests: %v", err)
		return
	}

	for _, pr := range pullreqList {
		if err = fn(pr); err != nil {
			log.Errorf("failed to process pull req: %v", err)
		}
	}
}

func getBranchFromRef(ref string) (string, error) {
	const refPrefix = "refs/heads/"
	if !strings.HasPrefix(ref, refPrefix) {
		return "", errors.Newf("failed to get branch name from branch ref %s", ref)
	}

	branch := ref[len(refPrefix):]
	if len(branch) == 0 {
		return "", errors.Newf("got an empty branch name from branch ref %s", ref)
	}
	return branch, nil
}
