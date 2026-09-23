// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import "time"

type Config struct {
	// InstanceID specifis the ID of the instance.
	// 实例ID指定了实例的ID。
	InstanceID string `envconfig:"INSTANCE_ID"`

	// MaxRunning is maximum number of jobs that can be running at once.
	// MaxRunning指定了同时可以运行的作业数量的最大值。
	BackgroundJobsMaxRunning int `envconfig:"JOBS_MAX_RUNNING" default:"10"`

	// RetentionTime is the duration after which non-recurring,
	// finished and failed jobs will be purged from the DB.
	// RetentionTime指定了非重复作业完成或失败后，从数据库中删除的持续时间。 默认值为5天。
	BackgroundJobsRetentionTime time.Duration `envconfig:"JOBS_RETENTION_TIME" default:"120h"` // 5 days

}
