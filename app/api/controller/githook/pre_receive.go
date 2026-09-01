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
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/githook"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	gitutil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/util"

	// "github.com/easysoft/gitfox/types"
	// "github.com/easysoft/gitfox/types/enum"

	"github.com/gotidy/ptr"
	"golang.org/x/exp/slices"
)

// PreReceive executes the pre-receive hook for a git repository.
// PreRestore执行git存储库的预接收挂钩。
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

	if !in.Internal && repo.Status != constant.RepoStateActive {
		output.Error = ptr.String("Push not allowed in the current repository state")
		return output, nil
	}

	// if err := c.limiter.RepoSize(ctx, in.RepoID); err != nil {
	// 	return hook.Output{}, fmt.Errorf("resource limit exceeded")
	// }

	forced := make([]bool, len(in.RefUpdates))
	for i, refUpdate := range in.RefUpdates {
		// 强制推送检测
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
	if !in.Internal && c.blockPullReqRefUpdate(refUpdates, repo.Status) {
		output.Error = ptr.String("pull reqRefs cantBeModified")
		return output, nil
	}

	// For internal calls - through the application interface (API) - no need to verify protection rules.
	if !in.Internal && repo.Status == constant.RepoStateActive {
		// TODO: use store.PrincipalInfoCache once we abstracted principals.
		_, err := c.userStore.FindByAccount(ctx, in.PrincipalAccount)
		if err != nil {
			return hook.Output{}, errors.Wrapf(err, "failed to find inner principal with account %s", in.PrincipalAccount)
		}

		err = c.checkProtectionRules(ctx, in.PrincipalAccount, repo, refUpdates, &output)
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

func (c *Controller) blockPullReqRefUpdate(refUpdates changedRefs, state string) bool {
	if state == constant.RepoStateImporting {
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

// 检查保护规则
func (c *Controller) checkProtectionRules(
	ctx context.Context,
	principalAccount string,
	repo *repoModel.Repo,
	refUpdates changedRefs,
	output *hook.Output,
) error {
	// 定义所有需要检查的分支操作
	checks := []struct {
		branches        []string
		actionType      constant.GitOperationType
		getAllowedUsers func(*utils.BranchPermissionResult) []string
		// enforceReview 表示该操作受强制评审(forceReview)限制
		enforceReview bool
	}{
		{refUpdates.branches.deleted, constant.GitOperationDelete, func(r *utils.BranchPermissionResult) []string { return r.DeleteUser }, false},
		{refUpdates.branches.created, constant.GitOperationCreate, func(r *utils.BranchPermissionResult) []string { return r.CreateUser }, false},
		{refUpdates.branches.updated, constant.GitOperationUpdate, func(r *utils.BranchPermissionResult) []string { return r.UpdateUser }, true},
		{refUpdates.branches.forced, constant.GitOperationForcePush, func(r *utils.BranchPermissionResult) []string { return r.ForcePushUser }, true},
	}

	for _, check := range checks {
		if len(check.branches) == 0 {
			continue
		}

		// 从数据库获取 DevOps Token
		entry, err := c.tokenStore.GetDevopsToken(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get devops token")
		}

		for _, branchName := range check.branches {
			rule, err := utils.GetBranchPermission(ctx, c.devopsConfig.URL, int64(repo.ID), branchName, entry.Key)
			if err != nil {
				return errors.Wrapf(err, "failed to get branch rule for %s", branchName)
			}

			// 强制评审开启时，禁止直接更新/强推该分支，必须通过 fit 提交
			if check.enforceReview && rule.ForceReview != 0 {
				output.Error = ptr.String(fmt.Sprintf(
					"branch %s must be use fit to Push",
					branchName))
				return nil
			}

			allowedUsers := check.getAllowedUsers(rule)
			// 如果允许的用户列表为空，表示没有限制
			if len(allowedUsers) == 0 {
				continue
			}

			// 检查当前用户是否在允许列表中
			if !slices.Contains(allowedUsers, principalAccount) {
				output.Error = ptr.String(fmt.Sprintf("not allowed to perform %s operation on branch %s", check.actionType, branchName))
				return nil
			}
		}
	}

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
