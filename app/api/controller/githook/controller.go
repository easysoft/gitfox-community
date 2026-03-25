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

	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Controller struct {
	userStore     user.Interface
	repoStore     repo.Interface
	pullreqStore  pullrequest.Interface
	tokenStore    token.Interface
	git           git.Interface
	urlProvider   url.Provider
	devopsConfig  config.DevopsConfig
	prReporter    *pullreqevents.Reporter
	gitReporter   *gitevents.Reporter
}

func NewController(
	userStore user.Interface,
	repoStore repo.Interface,
	pullreqStore pullrequest.Interface,
	tokenStore token.Interface,
	git git.Interface,
	urlProvider url.Provider,
	devopsConfig config.DevopsConfig,
	prReporter *pullreqevents.Reporter,
	gitReporter *gitevents.Reporter,
) *Controller {
	return &Controller{
		userStore:     userStore,
		repoStore:     repoStore,
		pullreqStore:  pullreqStore,
		tokenStore:    tokenStore,
		git:           git,
		urlProvider:   urlProvider,
		devopsConfig:  devopsConfig,
		prReporter:    prReporter,
		gitReporter:   gitReporter,
	}
}

func (c *Controller) getRepoCheckAccess(
	ctx context.Context,
	repoID int64,
) (*repoModel.Repo, error) {
	if repoID < 1 {
		return nil, errors.BadRequest.New("A valid repository reference must be provided.")
	}

	repo, err := c.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return nil, fmt.Errorf("failed to find repo with id %d: %w", repoID, err)
	}

	return repo, nil
}

// GetBaseSHAForScanningChanges returns the commit sha to which the new sha of the reference
// should be compared against when scanning incoming changes.
// NOTE: If no such a sha exists, then (sha.None, false, nil) is returned.
// This will happen in case the default branch doesn't exist yet.
func GetBaseSHAForScanningChanges(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *repoModel.Repo,
	env hook.Environment,
	refUpdates []hook.ReferenceUpdate,
	findBaseFor hook.ReferenceUpdate,
) (sha.SHA, bool, error) {
	// always return old SHA of ref if possible (even if ref was deleted, that's on the caller)
	if !findBaseFor.Old.IsNil() {
		return findBaseFor.Old, true, nil
	}

	// reference is just being created.
	// For now we use default branch as a fallback (can be optimized to most recent commit on reference that exists)
	dfltBranchFullRef := api.BranchPrefix + repo.DefaultBranch
	for _, refUpdate := range refUpdates {
		if refUpdate.Ref != dfltBranchFullRef {
			continue
		}

		// default branch is being updated as part of push - make sure we use OLD default branch sha for comparison
		if !refUpdate.Old.IsNil() {
			return refUpdate.Old, true, nil
		}

		// default branch is being created - no fallback available
		return sha.None, false, nil
	}

	// read default branch from git
	dfltBranchOut, err := rgit.GetBranch(ctx, &git.GetBranchParams{
		ReadParams: git.ReadParams{
			RepoUID:             repo.GitUID,
			AlternateObjectDirs: env.AlternateObjectDirs,
		},
		BranchName: repo.DefaultBranch,
	})
	if errors.IsNotFound(err) {
		// this happens for empty repo's where the default branch wasn't created yet.
		return sha.None, false, nil
	}
	if err != nil {
		return sha.None, false, fmt.Errorf("failed to get default branch from git: %w", err)
	}

	return dfltBranchOut.Branch.SHA, true, nil
}

func isForcePush(
	ctx context.Context,
	rgit RestrictedGIT,
	gitUID string,
	alternateObjectDirs []string,
	branchUpdate hook.ReferenceUpdate,
) (bool, error) {
	if branchUpdate.Old.IsNil() || branchUpdate.New.IsNil() {
		return false, nil
	}

	result, err := rgit.IsAncestor(ctx, git.IsAncestorParams{
		ReadParams: git.ReadParams{
			RepoUID:             gitUID,
			AlternateObjectDirs: alternateObjectDirs,
		},
		AncestorCommitSHA:   branchUpdate.Old,
		DescendantCommitSHA: branchUpdate.New,
	})
	if err != nil {
		return false, err
	}

	return !result.Ancestor, nil
}
