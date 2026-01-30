//go:build wireinject
// +build wireinject

package jobs

import (
	"time"

	"github.com/google/wire"

	jobsStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

// WireSet provides the wire set for job service.
var WireSet = wire.NewSet(
	ProvideScheduler,
	ProvideJobConfig,
)

// ProvideScheduler provides a job scheduler instance.
func ProvideScheduler(
	jobStore jobsStore.Interface,
	mxManager lock.MutexManager,
	pubsubService pubsub.PubSub,
	log *logger.SugaredMateLogger,
	cfg config.JobConfig,
) (*Scheduler, error) {
	return NewScheduler(
		jobStore,
		mxManager,
		pubsubService,
		log,
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

