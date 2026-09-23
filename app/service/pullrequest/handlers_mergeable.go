// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pullrequest

import (
	"context"
	"strconv"
	"time"

	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	gitenum "github.com/easysoft/gitfox-community/pkg/git/enum"
	"github.com/easysoft/gitfox-community/pkg/git/sha"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const (
	cancelMergeCheckKey = "cancel_merge_check_for_sha"
)

// mergeCheckOnCreated handles pull request Created events.
// It creates the PR head git ref.
func (s *Service) mergeCheckOnCreated(ctx context.Context,
	event *events.Event[*pullreqevents.CreatedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request created event for merge check, prID=%d", event.Payload.PullReqID)

	err := s.updateMergeData(
		ctx,
		event.Payload.TargetRepoID,
		event.Payload.PullReqID,
		sha.Nil.String(),
		event.Payload.SourceSHA,
	)
	if err != nil {
		return err
	}

	log.Infof("[event]: successfully consumed pull request created event for merge check, prID=%d", event.Payload.PullReqID)
	return nil
}

// mergeCheckOnBranchUpdate handles pull request Branch Updated events.
// It updates the PR head git ref to point to the latest commit.
func (s *Service) mergeCheckOnBranchUpdate(ctx context.Context,
	event *events.Event[*pullreqevents.BranchUpdatedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request branch updated event for merge check, prID=%d", event.Payload.PullReqID)

	err := s.updateMergeData(
		ctx,
		event.Payload.TargetRepoID,
		event.Payload.PullReqID,
		event.Payload.OldSHA,
		event.Payload.NewSHA,
	)
	if err != nil {
		return err
	}

	log.Infof("[event]: successfully consumed pull request branch updated event for merge check, prID=%d", event.Payload.PullReqID)
	return nil
}

// mergeCheckOnReopen handles pull request StateChanged events.
// It updates the PR head git ref to point to the source branch commit SHA.
func (s *Service) mergeCheckOnReopen(ctx context.Context,
	event *events.Event[*pullreqevents.ReopenedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request reopened event for merge check, prID=%d", event.Payload.PullReqID)

	err := s.updateMergeData(
		ctx,
		event.Payload.TargetRepoID,
		event.Payload.PullReqID,
		sha.None.String(),
		event.Payload.SourceSHA,
	)
	if err != nil {
		return err
	}

	log.Infof("[event]: successfully consumed pull request reopened event for merge check, prID=%d", event.Payload.PullReqID)
	return nil
}

// mergeCheckOnClosed deletes the merge ref.
func (s *Service) mergeCheckOnClosed(ctx context.Context,
	event *events.Event[*pullreqevents.ClosedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request closed event for merge check, prID=%d", event.Payload.PullReqID)

	err := s.deleteMergeRef(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID)
	if err != nil {
		return err
	}

	log.Infof("[event]: successfully consumed pull request closed event for merge check, prID=%d", event.Payload.PullReqID)
	return nil
}

// mergeCheckOnMerged deletes the merge ref.
func (s *Service) mergeCheckOnMerged(ctx context.Context,
	event *events.Event[*pullreqevents.MergedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request merged event for merge check, prID=%d", event.Payload.PullReqID)

	err := s.deleteMergeRef(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID)
	if err != nil {
		return err
	}

	log.Infof("[event]: successfully consumed pull request merged event for merge check, prID=%d", event.Payload.PullReqID)
	return nil
}

func (s *Service) deleteMergeRef(ctx context.Context, repoID int64, prID int64) error {
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return errors.Wrapf(err, "failed to get repo with ID %d", repoID)
	}

	writeParams, err := CreateSystemRPCWriteParams(ctx, s.urlProvider, int64(repo.ID), repo.GitUID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rpc write params")
	}

	// TODO: This doesn't work for forked repos
	err = s.gitSvc.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.FormatInt(prID, 10),
		Type:        gitenum.RefTypePullReqMerge,
		NewValue:    sha.None, // when NewValue is empty will delete the ref.
		OldValue:    sha.None, // we don't care about the old value
	})
	if err != nil {
		return errors.Wrapf(err, "failed to remove PR merge ref")
	}

	return nil
}

//nolint:funlen // refactor if required.
func (s *Service) updateMergeData(
	ctx context.Context,
	repoID int64,
	prID int64,
	oldSHA string,
	newSHA string,
) error {
	log := logger.GetLogger(ctx)

	pr, err := s.pullRequestStore.GetByID(ctx, uint(repoID), uint(prID))
	if err != nil {
		return errors.Wrapf(err, "failed to get pull request ID %d", prID)
	}

	if pr.Status != repoModel.PullRequestStatusOpen {
		return errors.Newf("cannot do mergability check on closed PR %d", pr.ID)
	}

	// load repository objects
	targetRepo, err := s.repoStore.GetByID(ctx, pr.TargetRepoID)
	if err != nil {
		return errors.Wrapf(err, "failed to get target repo")
	}

	sourceRepo := targetRepo
	if pr.TargetRepoID != pr.SourceRepoID {
		sourceRepo, err = s.repoStore.GetByID(ctx, pr.SourceRepoID)
		if err != nil {
			return errors.Wrapf(err, "failed to get source repo")
		}
	}

	writeParams, err := CreateSystemRPCWriteParams(ctx, s.urlProvider, int64(targetRepo.ID), targetRepo.GitUID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rpc write params")
	}

	// call merge and store output in pr merge reference.
	now := time.Now()
	mergeOutput, err := s.gitSvc.Merge(ctx, &git.MergeParams{
		WriteParams:     writeParams,
		BaseBranch:      pr.TargetBranch,
		HeadRepoUID:     sourceRepo.GitUID,
		HeadBranch:      pr.SourceBranch,
		RefType:         gitenum.RefTypePullReqMerge,
		RefName:         strconv.FormatInt(int64(pr.ID), 10),
		HeadExpectedSHA: sha.Must(newSHA),
		Force:           true,

		// set committer date to ensure repeatability of merge commit across replicas
		CommitterDate: &now,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to run git merge with base %q and head %q", pr.TargetBranch, pr.SourceBranch)
	}

	// Update DB in both cases (failure or success)
	_, err = s.pullRequestStore.UpdateOptLock(ctx, pr, func(pr *repoModel.PullRequest) error {
		// to avoid racing conditions with merge
		if pr.Status != repoModel.PullRequestStatusOpen {
			return errPRNotOpen
		}

		if pr.SourceSHA != newSHA {
			log.Warnf("PR SHA %s is newer than %s", pr.SourceSHA, newSHA)
			return nil
		}

		pr.MergeBaseSHA = mergeOutput.MergeBaseSHA.String()
		pr.MergeTargetSHA = mergeOutput.BaseSHA.String()
		if mergeOutput.MergeSHA.IsEmpty() {
			pr.MergeSHA = ""
		} else {
			pr.MergeSHA = mergeOutput.MergeSHA.String()
		}

		// Update merge check status based on conflict files
		if len(mergeOutput.ConflictFiles) > 0 {
			pr.MergeCheckStatus = repoModel.MergeCheckStatusConflict
			pr.MergeConflicts = mergeOutput.ConflictFiles
		} else {
			pr.MergeCheckStatus = repoModel.MergeCheckStatusMergeable
			pr.MergeConflicts = nil
		}

		// Update diff stats
		commitCount := uint(mergeOutput.CommitCount)
		fileCount := uint(mergeOutput.ChangedFileCount)
		additions := uint(mergeOutput.Additions)
		deletions := uint(mergeOutput.Deletions)
		pr.CommitCount = commitCount
		pr.FileCount = fileCount
		pr.Additions = additions
		pr.Deletions = deletions

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update PR merge ref in db")
	}

	if err = s.sseStreamer.Publish(ctx, int64(targetRepo.SpaceID), eventModel.SSETypePullRequestUpdated, pr); err != nil {
		log.Warnf("failed to publish PR changed event: %v", err)
	}

	// Avoid unused variable warning
	_ = oldSHA

	return nil
}
