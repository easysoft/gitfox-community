// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pullrequest

import (
	"context"
	"strconv"

	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/git"
	gitenum "github.com/easysoft/gitfox-community/pkg/git/enum"
	"github.com/easysoft/gitfox-community/pkg/git/sha"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// createHeadRefOnCreated handles pull request Created events.
// It creates the PR head git ref.
func (s *Service) createHeadRefOnCreated(ctx context.Context,
	event *events.Event[*pullreqevents.CreatedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request created event, prID=%d", event.Payload.PullReqID)

	repo, err := s.repoStore.GetByID(ctx, uint(event.Payload.TargetRepoID))
	if err != nil {
		return errors.Wrapf(err, "failed to get repo info")
	}

	writeParams, err := CreateSystemRPCWriteParams(ctx, s.urlProvider, int64(repo.ID), repo.GitUID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rpc write params")
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo).
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.gitSvc.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.FormatInt(event.Payload.PullReqID, 10),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    sha.Must(event.Payload.SourceSHA),
		OldValue:    sha.None, // this is a new pull request, so we expect that the ref doesn't exist
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update PR head ref")
	}

	log.Infof("[event]: successfully consumed pull request created event, prID=%d", event.Payload.PullReqID)
	return nil
}

// updateHeadRefOnBranchUpdate handles pull request Branch Updated events.
// It updates the PR head git ref to point to the latest commit.
func (s *Service) updateHeadRefOnBranchUpdate(ctx context.Context,
	event *events.Event[*pullreqevents.BranchUpdatedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request branch updated event, prID=%d", event.Payload.PullReqID)

	repo, err := s.repoStore.GetByID(ctx, uint(event.Payload.TargetRepoID))
	if err != nil {
		return errors.Wrapf(err, "failed to get repo info")
	}

	writeParams, err := CreateSystemRPCWriteParams(ctx, s.urlProvider, int64(repo.ID), repo.GitUID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rpc write params")
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo)
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.gitSvc.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.FormatInt(event.Payload.PullReqID, 10),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    sha.Must(event.Payload.NewSHA),
		OldValue:    sha.Must(event.Payload.OldSHA),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update PR head ref after new commit")
	}

	log.Infof("[event]: successfully consumed pull request branch updated event, prID=%d", event.Payload.PullReqID)
	return nil
}

// updateHeadRefOnReopen handles pull request StateChanged events.
// It updates the PR head git ref to point to the source branch commit SHA.
func (s *Service) updateHeadRefOnReopen(ctx context.Context,
	event *events.Event[*pullreqevents.ReopenedPayload],
) error {
	log := logger.GetLogger(ctx)
	log.Infof("[event]: consuming pull request reopened event, prID=%d", event.Payload.PullReqID)

	repo, err := s.repoStore.GetByID(ctx, uint(event.Payload.TargetRepoID))
	if err != nil {
		return errors.Wrapf(err, "failed to get repo info")
	}

	writeParams, err := CreateSystemRPCWriteParams(ctx, s.urlProvider, int64(repo.ID), repo.GitUID)
	if err != nil {
		return errors.Wrapf(err, "failed to generate rpc write params")
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo)
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.gitSvc.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.FormatInt(event.Payload.PullReqID, 10),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    sha.Must(event.Payload.SourceSHA),
		OldValue:    sha.None, // the request is re-opened, so anything can be the old value
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update PR head ref after pull request reopen")
	}

	log.Infof("[event]: successfully consumed pull request reopened event, prID=%d", event.Payload.PullReqID)
	return nil
}
