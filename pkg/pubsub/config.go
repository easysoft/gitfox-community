// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pubsub

import "time"

type Provider string

const (
	ProviderMemory Provider = "inmemory"
	ProviderRedis  Provider = "redis"
)

type Config struct {
	App       string
	Namespace string

	Provider Provider

	HealthInterval time.Duration
	SendTimeout    time.Duration
	ChannelSize    int
}
