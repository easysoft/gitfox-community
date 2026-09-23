// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package githook

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/app/model/githook"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/git/sha"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const (
	// gitReferenceNamePrefixBranch is the prefix of references of type branch.
	gitReferenceNamePrefixBranch = "refs/heads/"
	// gitReferenceNamePrefixTag is the prefix of references of type tag.
	gitReferenceNamePrefixTag = "refs/tags/"
	// gitReferenceNamePrefixTag is the prefix of pull req references.
	gitReferenceNamePullReq = "refs/pullreq/"
)

// PostReceive executes the post-receive hook for a git repository.
// PostReceive执行git存储库的post-receive钩子。
func (c *Controller) PostReceive(
	ctx context.Context,
	rgit RestrictedGIT,
	// session *auth.Session,
	in githook.GithookPostReceiveInput,
) (hook.Output, error) {
	repo, err := c.getRepoCheckAccess(ctx, in.RepoID)
	if err != nil {
		return hook.Output{}, err
	}
	// create output object and have following messages fill its messages
	out := hook.Output{}

	// report ref events if repo is in an active state (best effort)
	if repo.Status == constant.RepoStateActive {
		c.reportReferenceEvents(ctx, rgit, repo, in.PrincipalAccount, in.PostReceiveInput)
	}

	// handle branch updates related to PRs - best effort
	c.handlePRMessaging(ctx, repo, in.PrincipalAccount, in.PostReceiveInput, &out)

	// err = c.postReceiveExtender.Extend(ctx, rgit, session, repo, in, &out)
	// if err != nil {
	// 	return hook.Output{}, fmt.Errorf("failed to extend post-receive hook: %w", err)
	// }

	return out, nil
}

// reportReferenceEvents is reporting reference events to the event system.
// NOTE: keep best effort for now as it doesn't change the outcome of the git operation.
// TODO: in the future we might want to think about propagating errors so user is aware of events not being triggered.
func (c *Controller) reportReferenceEvents(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *repoModel.Repo,
	principalAccount string,
	in hook.PostReceiveInput,
) {
	for _, refUpdate := range in.RefUpdates {
		switch {
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixBranch):
			c.reportBranchEvent(ctx, rgit, repo, principalAccount, in.Environment, refUpdate, refUpdate.Ref, refUpdate.Old)
		case c.pushRequest != nil && c.pushRequest.IsPushRequestRef(refUpdate.Ref):
			ref, effectiveOld := c.pushRequest.ResolveReport(ctx, rgit, repo, in.Environment, refUpdate)
			c.reportBranchEvent(ctx, rgit, repo, principalAccount, in.Environment, refUpdate, ref, effectiveOld)
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixTag):
			c.reportTagEvent(ctx, repo, principalAccount, refUpdate)
		default:
			// Ignore any other references in post-receive
		}
	}
}

func (c *Controller) reportBranchEvent(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *repoModel.Repo,
	principalAccount string,
	env hook.Environment,
	branchUpdate hook.ReferenceUpdate,
	ref string,
	effOld sha.SHA,
) {
	log := logger.GetLogger(ctx)

	log.Debugf("post receive reporting branch event for ref %q(source: %q)", ref, branchUpdate.Ref)

	if c.gitReporter == nil {
		return
	}

	switch {
	case effOld.IsNil():
		// Branch created
		log.Debugf("branch %q created", ref)
		c.gitReporter.BranchCreated(ctx, &gitevents.BranchCreatedPayload{
			RepoID:  int64(repo.ID),
			Account: principalAccount,
			Ref:     ref,
			SHA:     branchUpdate.New.String(),
		})
	case branchUpdate.New.IsNil():
		// Branch deleted
		log.Debugf("branch %q deleted", ref)
		c.gitReporter.BranchDeleted(ctx, &gitevents.BranchDeletedPayload{
			RepoID:  int64(repo.ID),
			Account: principalAccount,
			Ref:     ref,
			SHA:     branchUpdate.Old.String(),
		})
	default:
		// Branch updated
		// A force update event might trigger some additional operations that aren't required
		// for ordinary updates (force pushes alter the commit history of a branch).
		forced, err := isForcePush(ctx, rgit, repo.GitUID, env.AlternateObjectDirs, branchUpdate)
		if err != nil {
			// In case of an error consider this a forced update. In post-update the branch has already been updated,
			// so there's less harm in declaring the update as forced.
			forced = true
			log.Warnf("failed to check ancestor for ref %s: %v", branchUpdate.Ref, err)
		}
		log.Debugf("branch %q updated, force push: %v", ref, forced)

		// best effort: fetch the commit message of the new head, an error here must not
		// block the event from being reported.
		var commitMessage string
		commitOut, err := rgit.GetCommit(ctx, &git.GetCommitParams{
			ReadParams: git.ReadParams{
				RepoUID:             repo.GitUID,
				AlternateObjectDirs: env.AlternateObjectDirs,
			},
			Revision: branchUpdate.New.String(),
		})
		if err != nil {
			log.Warnf("failed to get commit %s for ref %s: %v", branchUpdate.New.String(), branchUpdate.Ref, err)
		} else {
			commitMessage = commitOut.Commit.Message
		}

		c.gitReporter.BranchUpdated(ctx, &gitevents.BranchUpdatedPayload{
			RepoID:        int64(repo.ID),
			Account:       principalAccount,
			Ref:           ref,
			OldSHA:        effOld.String(),
			NewSHA:        branchUpdate.New.String(),
			Forced:        forced,
			CommitMessage: commitMessage,
		})
	}
}

func (c *Controller) reportTagEvent(
	ctx context.Context,
	repo *repoModel.Repo,
	principalAccount string,
	tagUpdate hook.ReferenceUpdate,
) {
	switch {
	case tagUpdate.Old.IsNil():
		c.gitReporter.TagCreated(ctx, &gitevents.TagCreatedPayload{
			RepoID:  int64(repo.ID),
			Account: principalAccount,
			Ref:     tagUpdate.Ref,
			SHA:     tagUpdate.New.String(),
		})
	case tagUpdate.New.IsNil():
		c.gitReporter.TagDeleted(ctx, &gitevents.TagDeletedPayload{
			RepoID:  int64(repo.ID),
			Account: principalAccount,
			Ref:     tagUpdate.Ref,
			SHA:     tagUpdate.Old.String(),
		})
	default:
		c.gitReporter.TagUpdated(ctx, &gitevents.TagUpdatedPayload{
			RepoID:  int64(repo.ID),
			Account: principalAccount,
			Ref:     tagUpdate.Ref,
			OldSHA:  tagUpdate.Old.String(),
			NewSHA:  tagUpdate.New.String(),
			Forced:  true,
		})
	}
}

// handlePRMessaging checks any single branch push for pr information and returns an according response if needed.
func (c *Controller) handlePRMessaging(
	ctx context.Context,
	repo *repoModel.Repo,
	principalAccount string,
	in hook.PostReceiveInput,
	out *hook.Output,
) {
	if len(in.RefUpdates) != 1 || in.RefUpdates[0].New.IsNil() {
		return
	}

	if c.pushRequest != nil && c.pushRequest.IsPushRequestRef(in.RefUpdates[0].Ref) {
		c.pushRequest.Handle(ctx, repo, principalAccount, in, out)
		return
	}

	branchName := in.RefUpdates[0].Ref[len(gitReferenceNamePrefixBranch):]
	c.suggestPullRequest(ctx, repo, branchName, out)
}

func (c *Controller) suggestPullRequest(
	ctx context.Context,
	repo *repoModel.Repo,
	branchName string,
	out *hook.Output,
) {
	log := logger.GetLogger(ctx)
	if branchName == repo.DefaultBranch {
		return
	}

	prs, err := c.pullreqStore.List(ctx, &repoModel.PullRequestFilter{
		PageSize:     2,
		SourceRepoID: repo.ID,
		SourceBranch: strings.TrimPrefix(branchName, "refs/heads/"),
		Status:       repoModel.PullRequestStatusOpen,
	})
	if err != nil {
		log.Warnf("failed to find pullrequests for branch '%s' originating from repo '%s'", branchName, repo.Path)
		return
	}

	if len(prs) > 0 {
		log.Debugf("found PRs for branch %s", branchName)
		msgs := make([]string, 2*len(prs)+1)
		msgs[0] = fmt.Sprintf("Branch %q has open PRs:", branchName)
		for i, pr := range prs {
			msgs[2*i+1] = fmt.Sprintf("  (#%d) %s", pr.ID, pr.Title)
			msgs[2*i+2] = "    " + c.urlProvider.GenerateUIPRURL(ctx, repo.Path, int64(pr.ID))
		}
		out.Messages = append(out.Messages, msgs...)
		return
	}

	out.Messages = append(out.Messages,
		fmt.Sprintf("Create a pull request for %q by visiting:", branchName),
		"  "+c.urlProvider.GenerateUICompareURL(ctx, repo.Path, repo.DefaultBranch, branchName),
	)
}

// eventBase creates a pullreqevents.Base from a PullRequest and account string.
func eventBase(pr *repoModel.PullRequest, account string) pullreqevents.Base {
	return pullreqevents.Base{
		PullReqID:    int64(pr.ID),
		SourceRepoID: int64(pr.SourceRepoID),
		TargetRepoID: int64(pr.TargetRepoID),
		Account:      account,
	}
}

// ReportPRCreated reports a pull request created event.
// This method is exposed for use when a PR is created externally.
func (c *Controller) ReportPRCreated(ctx context.Context, pr *repoModel.PullRequest, account string) {
	if c.prReporter == nil {
		return
	}
	c.prReporter.Created(ctx, &pullreqevents.CreatedPayload{
		Base:         eventBase(pr, account),
		SourceBranch: pr.SourceBranch,
		TargetBranch: pr.TargetBranch,
		SourceSHA:    pr.SourceSHA,
		MergeBaseSHA: pr.MergeBaseSHA,
	})
}

func strPtr(s string) *string {
	return &s
}
