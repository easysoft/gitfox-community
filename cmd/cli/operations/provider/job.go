// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package provider

import (
	"time"

	"github.com/google/wire"

	jobsSvc "github.com/easysoft/gitfox-community/app/service/jobs"
	jobsStore "github.com/easysoft/gitfox-community/app/store/jobs"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	runnerStore "github.com/easysoft/gitfox-community/app/store/runner"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/lock"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
)

// JobWireSet provides the wire set for job service.
var JobWireSet = wire.NewSet(
	jobsStore.NewScheduleJob,
	wire.Bind(new(jobsStore.Interface), new(*jobsStore.Store)),
	ProvideScheduler,
	wire.Bind(new(pipelineSvc.CronScheduler), new(*jobsSvc.Scheduler)),
	ProvideJobConfig,
)

// ProvideScheduler provides a job scheduler instance.
func ProvideScheduler(
	jobStore jobsStore.Interface,
	runnerStore runnerStore.Interface,
	mxManager lock.MutexManager,
	pubsubService pubsub.PubSub,
	jobLog *JobLog,
	cfg config.JobConfig,
) (*jobsSvc.Scheduler, error) {
	return jobsSvc.NewScheduler(
		jobStore,
		runnerStore,
		mxManager,
		pubsubService,
		jobLog.SugaredMateLogger,
		cfg.InstanceID,
		cfg.MaxRunning,
		cfg.RetentionTime,
	)
}

// ProvideJobConfig provides job scheduler configuration from system config.
func ProvideJobConfig(cfg *config.Config) config.JobConfig {
	instanceID := cfg.InstanceID
	if instanceID == "" {
		instanceID = "gitfox-default"
	}

	// Use config values if set, otherwise use defaults
	maxRunning := cfg.Job.MaxRunning
	if maxRunning < 1 {
		maxRunning = 10 // default max running jobs
	}

	retentionTime := cfg.Job.RetentionTime
	if retentionTime == 0 {
		retentionTime = 7 * 24 * time.Hour // 7 days
	}

	return config.JobConfig{
		InstanceID:    instanceID,
		MaxRunning:    maxRunning,
		RetentionTime: retentionTime,
	}
}
