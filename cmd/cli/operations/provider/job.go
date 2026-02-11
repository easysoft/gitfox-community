package provider

import (
	"time"

	"github.com/google/wire"

	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	jobsStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/jobs"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

// JobWireSet provides the wire set for job service.
var JobWireSet = wire.NewSet(
	jobsStore.NewScheduleJob,
	wire.Bind(new(jobsStore.Interface), new(*jobsStore.Store)),
	ProvideScheduler,
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
