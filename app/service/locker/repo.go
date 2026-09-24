// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package locker

import (
	"context"
	"strconv"
	"time"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

func (l Locker) LockDefaultBranch(ctx context.Context, repoID int64, branchName string, expiry time.Duration) (func(), error) {
	logger := logger.GetLogger(ctx)
	key := strconv.FormatInt(repoID, 10) + "/defaultBranch"

	logger.Info("attempting to lock to update the repo default branch")

	unlockFn, err := l.lock(ctx, namespaceRepo, key, expiry)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lock repo to update default branch to %s", branchName)
	}

	return unlockFn, nil
}
