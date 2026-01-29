package jobs

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
)

// globalLock 获取一个全局分布式锁
func globalLock(ctx context.Context, manager lock.MutexManager) (lock.Mutex, error) {
	const lockKey = "jobs"
	mx, err := manager.NewMutex(lockKey)
	if err != nil {
		return nil, err
	}

	err = mx.Lock(ctx)

	return mx, err
}
