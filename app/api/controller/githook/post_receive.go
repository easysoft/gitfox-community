package githook

import (
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/githook"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	gitenum "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

const (
	// gitReferenceNamePrefixBranch is the prefix of references of type branch.
	gitReferenceNamePrefixBranch = "refs/heads/"
	// gitReferenceNamePrefixTag is the prefix of references of type tag.
	gitReferenceNamePrefixTag = "refs/tags/"
	// gitReferenceNamePrefixTag is the prefix of pull req references.
	gitReferenceNamePullReq = "refs/pullreq/"
	// gitReferenceNameAgitPullReq special ref to create a pull request: refs/for/<targe-branch>/<topic-branch>
	// or refs/for/<targe-branch> -o topic='<topic-branch>'
	gitReferenceNameAgitPullReq = "refs/for/"
	// AGit Flow
)

// PullRequestFlow represents the type of pull request flow.
type PullRequestFlow string

const (
	PullRequestFlowGithub PullRequestFlow = "github"
	PullRequestFlowZentao PullRequestFlow = "zentao"
)

func (f PullRequestFlow) String() string {
	return string(f)
}

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
			c.reportBranchEvent(ctx, rgit, repo, principalAccount, in.Environment, refUpdate, in, PullRequestFlowGithub)
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNameAgitPullReq):
			c.reportBranchEvent(ctx, rgit, repo, principalAccount, in.Environment, refUpdate, in, PullRequestFlowZentao)
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
	in hook.PostReceiveInput,
	flow PullRequestFlow,
) {
	log := logger.GetLogger(ctx)

	ref := branchUpdate.Ref
	if flow == PullRequestFlowZentao {
		ref = strings.TrimPrefix(ref, gitReferenceNameAgitPullReq)
		ref = fmt.Sprintf("%s%s", c.getZtFlowBranchNamePre(ctx, repo, in), ref)
	}
	log.Debugf("post receive reporting branch event for ref %q(source: %q), %v", ref, branchUpdate.Ref, flow.String())

	if c.gitReporter == nil {
		return
	}

	switch {
	case branchUpdate.Old.IsNil():
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
			OldSHA:        branchUpdate.Old.String(),
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

	branchName := in.RefUpdates[0].Ref[len(gitReferenceNamePrefixBranch):]
	branchPre := ""
	if strings.HasPrefix(in.RefUpdates[0].Ref, gitReferenceNameAgitPullReq) {
		branchName = in.RefUpdates[0].Ref[len(gitReferenceNameAgitPullReq):]
		branchPre = c.getZtFlowBranchNamePre(ctx, repo, in)
	}
	c.suggestPullRequest(ctx, repo, branchName, branchPre, principalAccount, out)
}

func (c *Controller) suggestPullRequest(
	ctx context.Context,
	repo *repoModel.Repo,
	branchName, branchPre string,
	principalAccount string,
	out *hook.Output,
) {
	log := logger.GetLogger(ctx)
	if branchName == repo.DefaultBranch && len(branchPre) == 0 {
		return
	}

	if len(branchPre) > 0 {
		branchName = fmt.Sprintf("%s%s", branchPre, branchName)
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
		if strings.Contains(branchName, "/ztflow-") {
			bn := strings.TrimPrefix(branchName, branchPre)
			bn = strings.Split(bn, "#")[0]
			msgs[0] = fmt.Sprintf("Branch %q has open review PRs:", bn)
		}
		for i, pr := range prs {
			msgs[2*i+1] = fmt.Sprintf("  (#%d) %s", pr.ID, pr.Title)
			msgs[2*i+2] = "    " + c.urlProvider.GenerateUIPRURL(ctx, repo.Path, int64(pr.ID))
		}
		out.Messages = append(out.Messages, msgs...)
		return
	}

	log.Debugf("create new %s PR", branchName)
	if strings.Contains(branchName, "/ztflow-") {
		var sourceSHA string
		targetBranch := strings.TrimPrefix(branchName, branchPre)
		sourceBranch := strings.TrimPrefix(branchName, "refs/heads/")
		log.Debugf("branchName: %s, sourceBranch: %s, targetBranch: %s", branchName, sourceBranch, targetBranch)
		targetBranch = strings.Split(targetBranch, "#")[0]
		if sourceSHA, err = c.verifyBranchExistence(ctx, repo, sourceBranch); err != nil {
			log.Errorf("failed to verify source branch %s existence: %v", sourceBranch, err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		targetSHA, err := c.verifyBranchExistence(ctx, repo, targetBranch)
		if err != nil {
			log.Errorf("failed to verify target branch %s existence: %v", targetBranch, err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		if err = c.checkIfAlreadyExists(ctx, repo.ID, repo.ID, targetBranch, sourceBranch); err != nil {
			log.Errorf("check If Already Exists %s -> %s: %v", sourceBranch, targetBranch, err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		mergeBaseResult, err := c.git.MergeBase(ctx, git.MergeBaseParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			Ref1:       sourceBranch,
			Ref2:       targetBranch,
		})
		if err != nil {
			log.Errorf("failed to find merge base %s -> %s: %v", sourceBranch, targetBranch, err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		mergeBaseSHA := mergeBaseResult.MergeBaseSHA
		if mergeBaseSHA.String() == sourceSHA {
			log.Errorf("The source branch %s doesn't contain any new commits", sourceBranch)
			out.Error = strPtr("The source branch doesn't contain any new commits")
			return
		}
		// 查询分支权限获取 reviewFlowID
		entry, err := c.tokenStore.GetDevopsToken(ctx)
		if err != nil {
			log.Errorf("failed to get devops token: %v", err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		rule, err := utils.GetBranchPermission(ctx, c.devopsConfig.URL, int64(repo.ID), targetBranch, entry.Key)
		if err != nil {
			log.Errorf("failed to get branch rule for %s: %v", targetBranch, err)
			out.Error = strPtr("failed to create review pull request")
			return
		}
		pr := &repoModel.PullRequest{
			CreatedBy:         principalAccount,
			EditedBy:          principalAccount,
			RepoID:            repo.ID,
			SourceRepoID:      repo.ID,
			SourceBranch:      sourceBranch,
			TargetRepoID:      repo.ID,
			TargetBranch:      targetBranch,
			Status:            repoModel.PullRequestStatusOpen,
			MergeCheckStatus:  repoModel.MergeCheckStatusUnchecked,
			RebaseCheckStatus: repoModel.MergeCheckStatusUnchecked,
			Title:             fmt.Sprintf("发起 %q 强制评审", targetBranch),
			Desc:              "这是强制代码评审请求，请及时指定相关评审者",
			SourceSHA:         sourceSHA,
			MergeBaseSHA:      mergeBaseSHA.String(),
			MergeTargetSHA:    targetSHA,
			Flow:              repoModel.PullRequestFlowZentao,
			ReviewFlowID:      uint(rule.ReviewFlowID),
		}
		if err := c.pullreqStore.Create(ctx, pr); err != nil {
			out.Error = strPtr("failed to create a new pull request")
			return
		}
		if err := utils.SubmitCodeReview(ctx, c.devopsConfig.URL, int64(pr.ID), string(pr.CreatedBy), entry.Key); err != nil {
			log.Errorf("failed to submit code review for PPM %d: %v", pr.ID, err)
		}
		c.prReporter.Created(ctx, &pullreqevents.CreatedPayload{
			Base:         eventBase(pr, principalAccount),
			SourceBranch: sourceBranch,
			TargetBranch: targetBranch,
			SourceSHA:    sourceSHA,
			MergeBaseSHA: mergeBaseSHA.String(),
		})
		out.Messages = append(out.Messages,
			fmt.Sprintf("review a pull request for %q by visiting:", targetBranch),
			"  "+c.urlProvider.GenerateUIPRURL(ctx, repo.Path, int64(pr.ID)),
		)
		return
	}

	out.Messages = append(out.Messages,
		fmt.Sprintf("Create a pull request for %q by visiting:", branchName),
		"  "+c.urlProvider.GenerateUICompareURL(ctx, repo.Path, repo.DefaultBranch, branchName),
	)
}

func (c *Controller) checkIfAlreadyExists(ctx context.Context,
	targetRepoID, sourceRepoID uint, targetBranch, sourceBranch string,
) error {
	sourceBranch = strings.TrimPrefix(sourceBranch, "refs/heads/")
	existing, err := c.pullreqStore.List(ctx, &repoModel.PullRequestFilter{
		SourceRepoID: sourceRepoID,
		SourceBranch: sourceBranch,
		TargetRepoID: targetRepoID,
		TargetBranch: targetBranch,
		Status:       repoModel.PullRequestStatusOpen,
		PageSize:     1,
	})
	if err != nil {
		return fmt.Errorf("failed to get existing pull requests: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("a pull request for this target and source branch already exists")
	}

	return nil
}

func (c *Controller) verifyBranchExistence(ctx context.Context,
	repo *repoModel.Repo, branch string,
) (string, error) {
	if branch == "" {
		return "", errors.BadRequest.New("branch name can't be empty")
	}
	branch = strings.TrimPrefix(branch, "refs/heads/")
	ref, err := c.git.GetRef(ctx,
		git.GetRefParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			Name:       branch,
			Type:       gitenum.RefTypeBranch,
		})
	// if errors.AsStatus(err) == errors.StatusNotFound {
	// 	return "", usererror.BadRequest(
	// 		fmt.Sprintf("branch %q does not exist in the repository %q", branch, repo.Identifier))
	// }
	if err != nil {
		return "", fmt.Errorf(
			"failed to check existence of the branch %q in the repository %q: %w",
			branch, repo.Identifier, err)
	}

	return ref.SHA.String(), nil
}

func (c *Controller) getZtFlowBranchNamePre(ctx context.Context, repo *repoModel.Repo, in hook.PostReceiveInput) string {
	log := logger.GetLogger(ctx)
	branchPre := "refs/heads/ztflow-"
	if commit, err := c.git.GetCommit(ctx, &git.GetCommitParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		Revision:   in.RefUpdates[0].New.String(),
	}); err != nil {
		log.Errorf("failed to get commit %s: %v", in.RefUpdates[0].New.String(), err)
	} else {
		if commit.Commit.Author.Identity.Validate() == nil {
			email := strings.ToLower(commit.Commit.Author.Identity.Email)
			email = strings.Split(email, "@")[0]
			email = strings.ReplaceAll(email, ".", "")
			branchPre = fmt.Sprintf("%s%s-", branchPre, email)
		}
	}
	return branchPre
}

// func (c *Controller) notifyCodeOwners(ctx context.Context, session *auth.Session, repo *repoModel.Repo, pr *types.PullReq) {
// 	reviewers, err := c.reviewerStore.List(ctx, pr.ID)
// 	if err != nil {
// 		return
// 	}

// 	ownerEvaluation, err := c.codeOwners.Evaluate(ctx, repo, pr, reviewers)
// 	if err != nil {
// 		return
// 	}

// 	codeOwners := make(map[int64]types.PrincipalInfo)

// 	for _, entry := range ownerEvaluation.EvaluationEntries {
// 		for _, eva := range entry.OwnerEvaluations {
// 			codeOwners[eva.Owner.ID] = eva.Owner
// 		}
// 	}

// 	for _, owner := range codeOwners {
// 		c.prReporter.ReviewerAdded(ctx, &pullreqevents.ReviewerAddedPayload{
// 			Base:       eventBase(pr, &session.Principal),
// 			ReviewerID: owner.ID,
// 		})
// 	}
// }

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
