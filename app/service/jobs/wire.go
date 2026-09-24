// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

//go:build wireinject
// +build wireinject

package jobs

import (
	"time"

	"github.com/google/wire"

	jobsStore "github.com/easysoft/gitfox-community/app/store/jobs"
	runnerStore "github.com/easysoft/gitfox-community/app/store/runner"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/lock"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
)

// WireSet provides the wire set for job service.
var WireSet = wire.NewSet(
	ProvideScheduler,
	ProvideJobConfig,
)

// ProvideScheduler provides a job scheduler instance.
func ProvideScheduler(
	jobStore jobsStore.Interface,
	runnerStore runnerStore.Interface,
	mxManager lock.MutexManager,
	pubsubService pubsub.PubSub,
	log *logger.SugaredMateLogger,
	cfg config.JobConfig,
) (*Scheduler, error) {
	return NewScheduler(
		jobStore,
		runnerStore,
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
