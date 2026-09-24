// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package lock

import "time"

type Provider string

const (
	MemoryProvider Provider = "inmemory"
	RedisProvider  Provider = "redis"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration
type Config struct {
	App       string        // 应用名称前缀
	Namespace string        // 命名空间
	Provider  Provider      // 锁的实现提供者
	Expiry    time.Duration // 锁的过期时间

	Tries      int           // 最大重试次数
	RetryDelay time.Duration // 固定重试延迟
	DelayFunc  DelayFunc     // 自定义重试延迟函数

	DriftFactor   float64 // 漂移因子，用于调整锁的过期时间，以 account for clock skew
	TimeoutFactor float64 // 超时因子，用于调整锁的过期时间，以 account for clock skew

	GenValueFunc func() (string, error) // 生成锁值的函数
	Value        string                 // 自定义锁值
}
