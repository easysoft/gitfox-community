package locker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/easysoft/gitfox-community/pkg/errors"
)

func (l Locker) LockPR(ctx context.Context, repoID int64, prNum int64, expiry time.Duration) (func(), error) {
	key := fmt.Sprintf("%d/pulls", repoID)
	if prNum != 0 {
		key += "/" + strconv.FormatInt(prNum, 10)
	}

	unlockFn, err := l.lock(ctx, namespaceRepo, key, expiry)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lock mutex for pr %d in repo %d", prNum, repoID)
	}

	return unlockFn, nil
}
