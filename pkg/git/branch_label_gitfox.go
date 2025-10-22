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
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

type Label struct {
	Name  string
	Total int
}

func (s *Service) ListBranchLabelsV2(ctx context.Context, params *ListBranchesParams) ([]Label, error) {
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
		PageSize:      100000,
	})
	if err != nil {
		return nil, err
	}

	labelMap := make(map[string]int)
	for _, branch := range gitBranches {
		if strings.Contains(branch.Name, "ztflow-") {
			continue
		}

		splitName := strings.Split(strings.ToLower(branch.Name), "/")
		if len(splitName) > 1 {
			firstName := splitName[0]
			labelMap[firstName] = labelMap[firstName] + 1
		}
	}

	labels := make([]Label, len(labelMap))
	i := 0
	for k, v := range labelMap {
		labels[i] = Label{Name: k, Total: v}
		i++
	}

	return labels, nil
}
