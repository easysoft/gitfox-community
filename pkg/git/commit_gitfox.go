// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package git

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/git/api"
)

func (s *Service) ListCommitSHAs(ctx context.Context, params *ListCommitsParams) ([]string, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	gitCommits, err := s.git.ListCommitSHAs(
		ctx,
		repoPath,
		params.AlternateObjectDirs,
		params.GitREF,
		int(params.Page),
		int(params.Limit),
		api.CommitFilter{
			AfterRef:  params.After,
			Path:      params.Path,
			Since:     params.Since,
			Until:     params.Until,
			Committer: params.Committer,
			Author:    params.Author,
			Regex:     params.Regex,
		},
	)
	if err != nil {
		return nil, err
	}

	return gitCommits, nil
}
