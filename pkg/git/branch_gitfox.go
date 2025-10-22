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

package git

import (
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

func (s *Service) ListBranchesV2(ctx context.Context, params *ListBranchesParams) (*ListBranchesOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	// 默认查询所有分支
	gitBranches, err := s.listBranchesLoadReferenceData(ctx, repoPath, api.BranchFilter{
		IncludeCommit: params.IncludeCommit,
		Query:         "",
		Sort:          mapBranchesSortOption(params.Sort),
		Order:         mapToSortOrder(params.Order),
		Page:          1,
		PageSize:      10000,
	})
	if err != nil {
		return nil, err
	}

	var filteredBranches []*api.Branch
	for _, branch := range gitBranches {
		if strings.Contains(branch.Name, "ztflow-") {
			continue
		}
		if params.Query != "" && !strings.Contains(strings.ToLower(branch.Name), strings.ToLower(params.Query)) {
			continue
		}

		if !params.IncludeArchived && strings.HasPrefix(branch.Name, "archived/") {
			continue
		}

		names := strings.Split(strings.ToLower(branch.Name), "/")
		if params.Label != "" && (len(names) > 1 && params.Label != names[0] || len(names) == 1) {
			continue
		}
		filteredBranches = append(filteredBranches, branch)
	}

	// 应用分页
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start >= int32(len(filteredBranches)) {
		filteredBranches = []*api.Branch{}
	} else {
		if end > int32(len(filteredBranches)) {
			end = int32(len(filteredBranches))
		}
		filteredBranches = filteredBranches[start:end]
	}

	if params.IncludeCommit && len(filteredBranches) > 0 {
		commitSHAs := make([]string, len(filteredBranches))
		for i := range filteredBranches {
			commitSHAs[i] = filteredBranches[i].SHA.String()
		}

		gitCommits, err := s.git.GetCommits(ctx, repoPath, commitSHAs)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit: %w", err)
		}

		for i := range gitCommits {
			filteredBranches[i].Commit = gitCommits[i]
		}
	}

	branches := make([]Branch, len(filteredBranches))
	for i, branch := range filteredBranches {
		b, err := mapBranch(branch)
		if err != nil {
			return nil, err
		}
		branches[i] = *b
	}

	return &ListBranchesOutput{
		Branches: branches,
	}, nil
}
