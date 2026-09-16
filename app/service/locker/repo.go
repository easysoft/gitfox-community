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
