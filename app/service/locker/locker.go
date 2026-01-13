package locker

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

const namespaceRepo = "repo"

type Locker struct {
	mtxManager lock.MutexManager
}

func NewLocker(mtxManager lock.MutexManager) *Locker {
	return &Locker{
		mtxManager: mtxManager,
	}
}

func (l Locker) lock(ctx context.Context, namespace string, key string, expiry time.Duration) (func(), error) {
	logger := logger.GetLogger(ctx)

	// annotate logs for easier debugging of lock related issues
	ctx = context.WithValue(ctx, "key", key)
	ctx = context.WithValue(ctx, "namespace", namespaceRepo)
	ctx = context.WithValue(ctx, "expiry", expiry.String())

	mutext, err := l.mtxManager.NewMutex(
		key,
		lock.WithNamespace(namespace),
		lock.WithExpiry(expiry),
		lock.WithTimeoutFactor(4/expiry.Seconds()), // 4s
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new mutex")
	}

	logger.Debug("attempting to acquire lock")

	err = mutext.Lock(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lock the mutex")
	}

	logger.Debugf("successfully locked (expiry: %s)", expiry)

	unlockFn := func() {
		// always unlock independent of whether source context got canceled or not
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err := mutext.Unlock(ctx)
		if err != nil {
			logger.Warnf("failed to unlock: %v", err)
		} else {
			logger.Debug("successfully unlocked")
		}
	}

	return unlockFn, nil
}
