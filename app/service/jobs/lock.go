// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/lock"
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
