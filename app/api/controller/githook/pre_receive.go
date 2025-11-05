// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package githook

import (
	"context"
	"fmt"
	"strings"

	// apiauth "github.com/easysoft/gitfox/app/api/auth"
	// "github.com/easysoft/gitfox/app/api/controller/limiter"
	// "github.com/easysoft/gitfox/app/api/usererror"
	// "github.com/easysoft/gitfox/app/auth"
	// "github.com/easysoft/gitfox/app/services/protection"
	// "github.com/easysoft/gitfox/app/api/controller/limiter"
	// "github.com/easysoft/gitfox/app/api/usererror"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	gitutil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"

	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	// "github.com/easysoft/gitfox/types"
	// "github.com/easysoft/gitfox/types/enum"

	"github.com/gotidy/ptr"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

// PreReceive executes the pre-receive hook for a git repository.
func (c *Controller) PreReceive(
	ctx context.Context,
	rgit RestrictedGIT,
	// session *auth.Session,
	in githook.GithookPreReceiveInput,
) (hook.Output, error) {
	output := hook.Output{}
	repo, err := c.getRepoCheckAccess(ctx, in.RepoID)
	if err != nil {
		return hook.Output{}, err
	}

	if !in.Internal && repo.State != enum.RepoStateActive && repo.State != enum.RepoStateMigrateGitPush {
		output.Error = ptr.String("Push not allowed in the current repository state")
		return output, nil
	}

	// if err := c.limiter.RepoSize(ctx, in.RepoID); err != nil {
	// 	return hook.Output{}, fmt.Errorf("resource limit exceeded")
	// }

	forced := make([]bool, len(in.RefUpdates))
	for i, refUpdate := range in.RefUpdates {
		forced[i], err = isForcePush(
			ctx, rgit, repo.GitUID, in.Environment.AlternateObjectDirs, refUpdate,
		)
		if err != nil {
			return hook.Output{}, fmt.Errorf("failed to check branch ancestor: %w", err)
		}
	}

	refUpdates := groupRefsByAction(in.RefUpdates, forced)

	if slices.Contains(refUpdates.branches.deleted, repo.DefaultBranch) {
		// Default branch mustn't be deleted.
		output.Error = ptr.String("default branch can not delete")
		return output, nil
	}

	// For external calls (git pushes) block modification of pullreq references.
	if !in.Internal && c.blockPullReqRefUpdate(refUpdates, repo.State) {
		output.Error = ptr.String("pull reqRefs cantBeModified")
		return output, nil
	}

	// For internal calls - through the application interface (API) - no need to verify protection rules.
	if !in.Internal && repo.State == enum.RepoStateActive {
		// TODO: use store.PrincipalInfoCache once we abstracted principals.
		_, err := c.principalStore.Find(ctx, in.PrincipalID)
		if err != nil {
			return hook.Output{}, fmt.Errorf("failed to find inner principal with id %d: %w", in.PrincipalID, err)
		}

		// dummySession := &auth.Session{Principal: *principal, Metadata: nil}

		err = c.checkProtectionRules(ctx, repo, refUpdates, &output)
		if output.Error != nil {
			return output, nil
		}
		if err != nil {
			return hook.Output{}, fmt.Errorf("failed to check protection rules: %w", err)
		}
	}

	// err = c.ztFlow(ctx, rgit, repo, in, &output)
	// if output.Error != nil {
	// 	return output, nil
	// }
	// if err != nil {
	// 	log.Ctx(ctx).Error().Err(err).Msg("failed to execute ztflow")
	// 	return hook.Output{}, err
	// }

	// err = c.checkFileSizeLimit(ctx, rgit, repo, in, &output)
	// if output.Error != nil {
	// 	return output, nil
	// }
	// if err != nil {
	// 	return hook.Output{}, err
	// }

	//err = c.checkPushUpdateLines(ctx, rgit, repo, in, &output)
	//log.Ctx(ctx).Err(err).Msg("failed to check push update lines")
	//if output.Error != nil {
	//	return output, nil
	//}
	//if err != nil {
	//	log.Ctx(ctx).Err(err).Msg("failed to check push update lines")
	//	return hook.Output{}, err
	//}

	return output, nil
}

func (c *Controller) blockPullReqRefUpdate(refUpdates changedRefs, state enum.RepoState) bool {
	if state == enum.RepoStateMigrateGitPush {
		return false
	}

	fn := func(ref string) bool {
		return strings.HasPrefix(ref, gitReferenceNamePullReq)
	}

	return slices.ContainsFunc(refUpdates.other.created, fn) ||
		slices.ContainsFunc(refUpdates.other.deleted, fn) ||
		slices.ContainsFunc(refUpdates.other.updated, fn) ||
		slices.ContainsFunc(refUpdates.other.forced, fn)
}

func (c *Controller) checkProtectionRules(
	ctx context.Context,
	// session *auth.Session,
	repo *types.Repository,
	refUpdates changedRefs,
	output *hook.Output,
) error {
	// isRepoOwner, err := apiauth.IsRepoOwner(ctx, c.authorizer, nil, repo)
	// if err != nil {
	// 	return fmt.Errorf("failed to determine if user is repo owner: %w", err)
	// }

	// protectionRules, err := c.protectionManager.ForRepository(ctx, repo.ID)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch protection rules for the repository: %w", err)
	// }

	// var ruleViolations []types.RuleViolations
	// var errCheckAction error

	//nolint:unparam
	// checkAction := func(refAction protection.RefAction, refType protection.RefType, names []string) {
	// 	if errCheckAction != nil || len(names) == 0 {
	// 		return
	// 	}

	// 	violations, err := protectionRules.RefChangeVerify(ctx, protection.RefChangeVerifyInput{
	// 		Actor:       &session.Principal,
	// 		AllowBypass: true,
	// 		IsRepoOwner: isRepoOwner,
	// 		Repo:        repo,
	// 		RefAction:   refAction,
	// 		RefType:     refType,
	// 		RefNames:    names,
	// 	})
	// 	if err != nil {
	// 		errCheckAction = fmt.Errorf("failed to verify protection rules for git push: %w", err)
	// 		return
	// 	}

	// 	ruleViolations = append(ruleViolations, violations...)
	// }

	// checkAction(protection.RefActionCreate, protection.RefTypeBranch, refUpdates.branches.created)
	// checkAction(protection.RefActionDelete, protection.RefTypeBranch, refUpdates.branches.deleted)
	// checkAction(protection.RefActionUpdate, protection.RefTypeBranch, refUpdates.branches.updated)
	// checkAction(protection.RefActionUpdateForce, protection.RefTypeBranch, refUpdates.branches.forced)

	// if errCheckAction != nil {
	// 	return errCheckAction
	// }

	// var criticalViolation bool

	// for _, ruleViolation := range ruleViolations {
	// 	criticalViolation = criticalViolation || ruleViolation.IsCritical()
	// 	for _, violation := range ruleViolation.Violations {
	// 		var message string
	// 		if ruleViolation.Bypassed {
	// 			message = fmt.Sprintf("Bypassed rule %q: %s", ruleViolation.Rule.Identifier, violation.Message)
	// 		} else {
	// 			message = fmt.Sprintf("Rule %q violation: %s", ruleViolation.Rule.Identifier, violation.Message)
	// 		}
	// 		output.Messages = append(output.Messages, message)
	// 	}
	// }

	// if criticalViolation {
	// 	output.Error = ptr.String("Blocked by protection rules.")
	// }

	return nil
}

type changes struct {
	created []string
	deleted []string
	updated []string
	forced  []string
}

func (c *changes) groupByAction(
	refUpdate hook.ReferenceUpdate,
	name string,
	forced bool,
) {
	switch {
	case refUpdate.Old.IsNil():
		c.created = append(c.created, name)
	case refUpdate.New.IsNil():
		c.deleted = append(c.deleted, name)
	case forced:
		c.forced = append(c.forced, name)
	default:
		c.updated = append(c.updated, name)
	}
}

type changedRefs struct {
	branches changes
	tags     changes
	other    changes
}

func groupRefsByAction(refUpdates []hook.ReferenceUpdate, forced []bool) (c changedRefs) {
	for i, refUpdate := range refUpdates {
		switch {
		// Branches refs/heads
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixBranch):
			branchName := refUpdate.Ref[len(gitReferenceNamePrefixBranch):]
			c.branches.groupByAction(refUpdate, branchName, forced[i])
		// Tags refs/tags
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixTag):
			tagName := refUpdate.Ref[len(gitReferenceNamePrefixTag):]
			c.tags.groupByAction(refUpdate, tagName, false)
		// AGit Pull requests refs/for/
		case gitutil.SupportProcReceive() && strings.HasPrefix(refUpdate.Ref, gitReferenceNameAgitPullReq):
			branchName := refUpdate.Ref[len(gitReferenceNameAgitPullReq):]
			branchName = fmt.Sprintf("ztflow-%s", branchName)
			c.branches.groupByAction(refUpdate, branchName, forced[i])
		default:
			c.other.groupByAction(refUpdate, refUpdate.Ref, false)
		}
	}
	return
}

func loggingWithRefUpdate(refUpdate hook.ReferenceUpdate) func(c zerolog.Context) zerolog.Context {
	return func(c zerolog.Context) zerolog.Context {
		return c.Str("ref", refUpdate.Ref).Str("old_sha", refUpdate.Old.String()).Str("new_sha", refUpdate.New.String())
	}
}
