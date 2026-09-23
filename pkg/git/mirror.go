// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package git

import (
	"context"
)

type MirrorSyncParams struct {
	ReadParams
	Prune    bool
	CloneURL string
}

func (s *Service) MirrorSyncRepository(ctx context.Context, params *MirrorSyncParams) (bool, error) {
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	err, sync := s.git.MirrorFetch(ctx, repoPath, "", params.Prune)
	return sync, err
}
