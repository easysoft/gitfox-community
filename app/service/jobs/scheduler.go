package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	jobModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	jobsStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/jobs"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"github.com/gorhill/cronexpr"
)

const (
	PubSubTopicCancelJob   = "gitfox:job:cancel_job"
	PubSubTopicStateChange = "gitfox:job:state_change"

	ProgressMin = 0
	ProgressMax = 100
)

// Scheduler controls execution of background jobs.
type Scheduler struct {
	// dependencies 依赖项
	jobStore      jobsStore.Interface
	runnerStore   runnerStore.Interface
	executor      *pkgJobs.Executor
	mxManager     lock.MutexManager
	pubsubService pubsub.PubSub
	log           *logger.SugaredMateLogger

	// configuration fields 配置字段
	instanceID    string
	maxRunning    int
	retentionTime time.Duration

	// synchronization stuff 同步相关内容
	signal       chan time.Time
	done         chan struct{}
	wgRunning    sync.WaitGroup
	cancelJobMx  sync.Mutex
	cancelJobMap map[string]context.CancelFunc
}

func NewScheduler(
	jobStore jobsStore.Interface,
	runnerStore runnerStore.Interface,
	mxManager lock.MutexManager,
	pubsubService pubsub.PubSub,
	log *logger.SugaredMateLogger,
	instanceID string,
	maxRunning int,
	retentionTime time.Duration,
) (*Scheduler, error) {
	if maxRunning < 1 {
		maxRunning = 1
	}

	executor := pkgJobs.NewExecutor(pubsubService)

	return &Scheduler{
		jobStore:      jobStore,
		runnerStore:   runnerStore,
		executor:      executor,
		mxManager:     mxManager,
		pubsubService: pubsubService,
		log:           log,

		instanceID:    instanceID,
		maxRunning:    maxRunning,
		retentionTime: retentionTime,

		cancelJobMap: map[string]context.CancelFunc{},
	}, nil
}

// Run runs the background job scheduler.
// It's a blocking call. It blocks until the provided context is done.
//
//nolint:gocognit // refactor if needed.
func (s *Scheduler) Run(ctx context.Context) error {
	if s.done != nil {
		return errors.New("scheduler already started")
	}

	// Subscribe to cancel job topic
	consumer := s.pubsubService.Subscribe(ctx, PubSubTopicCancelJob, s.handleCancelJob)
	defer func() {
		err := consumer.Close()
		if err != nil {
			s.log.Errorw("[JobScheduler] failed to close pubsub cancel job consumer", "error", err)
		}
	}()

	// Register internal system jobs (purge, overdue)
	if err := s.registerSystemJobs(); err != nil {
		return fmt.Errorf("failed to register system jobs: %w", err)
	}

	// Create necessary system jobs in database
	if err := s.createSystemJobs(ctx); err != nil {
		return fmt.Errorf("failed to create system jobs: %w", err)
	}

	s.executor.FinishRegistration()

	s.log.Info("[JobScheduler] Starting job scheduler")

	s.done = make(chan struct{})
	defer close(s.done)

	s.signal = make(chan time.Time, 1)

	timer := pkgJobs.NewSchedulerTimer()
	defer timer.Stop()

	for {
		err := func() error {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					s.log.Errorw("[JobScheduler] panic recovered",
						"panic", fmt.Sprintf("[%T] job scheduler panic: %v", r, r),
						"stack", stack)
				}
			}()

			select {
			case <-ctx.Done():
				s.log.Info("[JobScheduler] Context done, stopping scheduler")
				return ctx.Err()

			case newTime := <-s.signal:
				dur := timer.RescheduleEarlier(newTime)
				if dur > 0 {
					s.log.Debugf("[JobScheduler] Update scheduled job processing time... runs in %s", dur)
				}
				return nil

			case now := <-timer.Ch():
				count, nextExec, gotAllJobs, err := s.processReadyJobs(ctx, now)

				// If the next processing time isn't known use the default.
				if nextExec.IsZero() {
					const period = time.Minute
					nextExec = now.Add(period)
				}

				// Reset the timer. Make the timer edgy if there are more jobs available.
				dur := timer.ResetAt(nextExec, !gotAllJobs)

				if err != nil {
					s.log.Errorw("[JobScheduler] failed to process jobs",
						"error", err,
						"next_iteration", dur.String())
				} else {
					s.log.Debugf("[JobScheduler] started %d jobs; next iteration in %s", count, dur)
				}

				return nil
			}
		}()
		if err != nil {
			return err
		}
	}
}

// WaitJobsDone waits until execution of all jobs has finished.
// It is intended to be used for graceful shutdown, after the Run method has finished.
func (s *Scheduler) WaitJobsDone(ctx context.Context) {
	s.log.Debug("[JobScheduler] Waiting for running jobs to finish")

	ch := make(chan struct{})
	go func() {
		s.wgRunning.Wait()
		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.log.Warn("[JobScheduler] Stop interrupted")
	case <-ch:
		s.log.Info("[JobScheduler] Gracefully stopped")
	}
}

// CancelJob cancels a currently running or scheduled job.
func (s *Scheduler) CancelJob(ctx context.Context, jobName string) error {
	mx, err := globalLock(ctx, s.mxManager)
	if err != nil {
		return fmt.Errorf("failed to obtain global lock to cancel a job: %w", err)
	}

	defer func() {
		if err := mx.Unlock(ctx); err != nil {
			s.log.Errorw("[JobScheduler] failed to release global lock after canceling a job", "error", err)
		}
	}()

	job, err := s.jobStore.GetByName(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to find job to cancel: %w", err)
	}

	if job.IsRecurring {
		return errors.New("can't cancel recurring jobs")
	}

	if job.State != jobModel.StateScheduled && job.State != jobModel.StateRunning {
		return nil // return no error if the job is already canceled or has finished or failed.
	}

	// first we update the job in the database...

	job.EditedDate = datetime.NowLocalTime()
	job.State = jobModel.StateCanceled

	err = s.jobStore.UpdateExecution(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update job to cancel it: %w", err)
	}

	// ... and then we cancel its context.

	s.cancelJobMx.Lock()
	cancelFn, ok := s.cancelJobMap[jobName]
	s.cancelJobMx.Unlock()

	if ok {
		cancelFn()
		return nil
	}

	return s.pubsubService.Publish(ctx, PubSubTopicCancelJob, []byte(jobName))
}

func (s *Scheduler) handleCancelJob(payload []byte) error {
	jobName := string(payload)
	if jobName == "" {
		return nil
	}

	s.cancelJobMx.Lock()
	cancelFn, ok := s.cancelJobMap[jobName]
	s.cancelJobMx.Unlock()

	if ok {
		cancelFn()
	}

	return nil
}

// scheduleProcessing triggers processing of ready jobs.
// This should be run after adding new jobs to the database.
func (s *Scheduler) scheduleProcessing(scheduled time.Time) {
	go func() {
		select {
		case <-s.done:
		case s.signal <- scheduled:
		}
	}()
}

// scheduleIfHaveMoreJobs triggers processing of ready jobs if the timer is edgy.
// The timer would be edgy if the previous iteration found more jobs that it could start (full capacity).
// This should be run after a non-recurring job has finished.
func (s *Scheduler) scheduleIfHaveMoreJobs() {
	s.scheduleProcessing(time.Time{}) // zero time will trigger the timer if it's edgy
}

// processReadyJobs executes jobs that are ready to run. This function is periodically run by the Scheduler.
// The function returns the number of jobs it has started, the next scheduled execution time (of this function)
// and a bool value if all currently available ready jobs were started.
func (s *Scheduler) processReadyJobs(ctx context.Context, now time.Time) (int, time.Time, bool, error) {
	mx, err := globalLock(ctx, s.mxManager)
	if err != nil {
		return 0, time.Time{}, false,
			fmt.Errorf("failed to obtain global lock to periodically process ready jobs: %w", err)
	}

	defer func() {
		if err := mx.Unlock(ctx); err != nil {
			s.log.Errorw("[JobScheduler] failed to release global lock after periodic processing of ready jobs",
				"error", err)
		}
	}()

	availableCount, err := s.availableSlots(ctx)
	if err != nil {
		return 0, time.Time{}, false,
			fmt.Errorf("failed to count available slots for job execution: %w", err)
	}

	// get one over the limit to check if all ready jobs are fetched
	jobs, err := s.jobStore.ListReady(ctx, now, availableCount+1)
	if err != nil {
		return 0, time.Time{}, false,
			fmt.Errorf("failed to load scheduled jobs: %w", err)
	}

	var (
		countExecuted     int
		knownNextExecTime time.Time
		gotAllJobs        bool
	)

	if len(jobs) > availableCount {
		// More jobs are ready than we are able to run.
		jobs = jobs[:availableCount]
	} else {
		gotAllJobs = true
		knownNextExecTime, err = s.jobStore.NextScheduledTime(ctx, now)
		if err != nil {
			return 0, time.Time{}, false,
				fmt.Errorf("failed to read next scheduled time: %w", err)
		}
	}

	for _, job := range jobs {
		jobLog := s.log.With("job.Name", job.Name, "job.Type", job.Type)

		// Update the job fields for the new execution
		s.preExec(job)

		if err := s.jobStore.UpdateExecution(ctx, job); err != nil {
			knownNextExecTime = time.Time{}
			gotAllJobs = false
			jobLog.Errorw("[JobScheduler] failed to update job to mark it as running", "error", err)
			continue
		}

		// tell everybody that a job has started
		if err := publishStateChange(ctx, s.pubsubService, job); err != nil {
			jobLog.Errorw("[JobScheduler] failed to publish job state change", "error", err)
		}

		s.runJob(ctx, job, jobLog)

		countExecuted++
	}

	return countExecuted, knownNextExecTime, gotAllJobs, nil
}

func (s *Scheduler) availableSlots(ctx context.Context) (int, error) {
	countRunning, err := s.jobStore.CountRunning(ctx)
	if err != nil {
		return 0, err
	}

	availableCount := s.maxRunning - countRunning
	if availableCount < 0 {
		return 0, nil
	}

	return availableCount, nil
}

// runJob updates the job in the database and starts it in a separate goroutine.
// The function will also log the execution.
func (s *Scheduler) runJob(ctx context.Context, j *jobModel.ScheduleJob, jobLog *logger.SugaredMateLogger) {
	s.wgRunning.Add(1)
	go func(ctx context.Context,
		jobName, jobType, jobData string,
		jobRunDeadline *datetime.Local,
		jobLog *logger.SugaredMateLogger,
	) {
		defer s.wgRunning.Done()

		jobLog.Debug("[JobScheduler] started job")

		timeStart := time.Now()

		// Run the job
		execResult, execFailure := s.doExec(ctx, jobName, jobType, jobData, jobRunDeadline)

		// Use the context.Background() because we want to update the job even if the job's context is done.
		// The context can be done because the job exceeded its deadline or the server is shutting down.
		backgroundCtx := context.Background()

		if mx, err := globalLock(backgroundCtx, s.mxManager); err != nil {
			// If locking failed, just log the error and proceed to update the DB anyway.
			jobLog.Errorw("[JobScheduler] failed to obtain global lock to update job after execution", "error", err)
		} else {
			defer func() {
				if err := mx.Unlock(backgroundCtx); err != nil {
					jobLog.Errorw("[JobScheduler] failed to release global lock to update job after execution", "error", err)
				}
			}()
		}

		job, err := s.jobStore.GetByName(backgroundCtx, jobName)
		if err != nil {
			jobLog.Errorw("[JobScheduler] failed to find job after execution", "error", err)
			return
		}

		// Update the job fields, reschedule if necessary.
		postExec(job, execResult, execFailure)

		err = s.jobStore.UpdateExecution(backgroundCtx, job)
		if err != nil {
			jobLog.Errorw("[JobScheduler] failed to update job after execution", "error", err)
			return
		}

		duration := time.Since(timeStart).String()
		logFields := []interface{}{"duration", duration}

		if job.IsRecurring {
			logFields = append(logFields, "job.IsRecurring", true)
		}
		if job.Result != "" {
			logFields = append(logFields, "job.Result", job.Result)
		}
		if job.LastFailureError != "" {
			logFields = append(logFields, "job.Failure", job.LastFailureError)
		}

		switch job.State {
		case jobModel.StateFinished:
			jobLog.Infow("[JobScheduler] job successfully finished", logFields...)
			s.scheduleIfHaveMoreJobs()

		case jobModel.StateFailed:
			jobLog.Infow("[JobScheduler] job failed", logFields...)
			s.scheduleIfHaveMoreJobs()

		case jobModel.StateCanceled:
			jobLog.Error("[JobScheduler] job canceled")
			s.scheduleIfHaveMoreJobs()

		case jobModel.StateScheduled:
			scheduledTime := time.Time(*job.Scheduled)
			logFields = append(logFields, "job.Scheduled", scheduledTime.Format(time.RFC3339Nano))
			jobLog.Infow("[JobScheduler] job finished and rescheduled", logFields...)

			s.scheduleProcessing(scheduledTime)

		case jobModel.StateRunning:
			jobLog.Error("[JobScheduler] should not happen; job still has state=running after finishing")
		}

		// tell everybody that a job has finished execution
		if err := publishStateChange(backgroundCtx, s.pubsubService, job); err != nil {
			jobLog.Errorw("[JobScheduler] failed to publish job state change", "error", err)
		}
	}(ctx, j.Name, j.Type, j.Data, j.RunDeadline, jobLog)
}

// preExec updates the provided Job before execution.
func (s *Scheduler) preExec(job *jobModel.ScheduleJob) {
	if job.MaxDurationSeconds < 1 {
		job.MaxDurationSeconds = 1
	}

	now := time.Now()
	nowLocal := datetime.Local(now)

	execDuration := time.Duration(job.MaxDurationSeconds) * time.Second
	execDeadline := now.Add(execDuration)
	execDeadlineLocal := datetime.Local(execDeadline)

	job.EditedDate = datetime.NowLocalTime()
	job.LastExecuted = &nowLocal
	job.State = jobModel.StateRunning
	job.RunDeadline = &execDeadlineLocal
	job.RunBy = s.instanceID
	job.RunProgress = ProgressMin
	job.TotalExecutions++
	job.Result = ""
	job.LastFailureError = ""
}

// doExec executes the provided Job.
func (s *Scheduler) doExec(ctx context.Context,
	jobName, jobType, jobData string,
	jobRunDeadline *datetime.Local,
) (execResult, execError string) {
	var execDeadline time.Time
	if jobRunDeadline != nil {
		execDeadline = time.Time(*jobRunDeadline)
	} else {
		execDeadline = time.Now().Add(30 * time.Second)
	}

	jobCtx, done := context.WithDeadline(ctx, execDeadline)
	defer done()

	s.cancelJobMx.Lock()
	if _, ok := s.cancelJobMap[jobName]; ok {
		// should not happen: jobs have unique names!
		s.cancelJobMx.Unlock()
		return "", "failed to start: already running"
	}
	s.cancelJobMap[jobName] = done
	s.cancelJobMx.Unlock()

	defer func() {
		s.cancelJobMx.Lock()
		delete(s.cancelJobMap, jobName)
		s.cancelJobMx.Unlock()
	}()

	progressReporter := func(progress int, result string) ([]byte, error) {
		if progress < ProgressMin || progress > ProgressMax {
			return nil, errors.New("progress must be between 0 and 100")
		}

		jobProgress := &jobModel.ScheduleJob{
			Name:        jobName,
			Type:        jobType,
			State:       jobModel.StateRunning,
			RunProgress: progress,
			EditedDate:  datetime.NowLocalTime(),
			Result:      result,
		}

		if err := s.jobStore.UpdateProgress(ctx, jobProgress); err != nil {
			return nil, err
		}

		// 发布进度更新事件，让订阅者能实时获取进度
		if err := publishStateChange(ctx, s.pubsubService, jobProgress); err != nil {
			s.log.Warnw("[JobScheduler] failed to publish job progress change",
				"job.Name", jobName,
				"job.Type", jobType,
				"progress", progress,
				"error", err)
			// 不返回错误，进度发布失败不应影响任务执行
		}

		return []byte(result), nil
	}

	execResult, err := s.executor.Exec(jobCtx, jobName, jobType, jobData, progressReporter)
	if err != nil {
		execError = err.Error()
	}

	return
}

// postExec updates the provided Job after execution and reschedules it if necessary.
//
//nolint:gocognit // refactor if needed.
func postExec(job *jobModel.ScheduleJob, resultData, resultErr string) {
	// Proceed with the update of the job if it's in the running state or
	// if it's marked as canceled but has succeeded nonetheless.
	// Other states should not happen, but if they do, just leave the job as it is.
	if job.State != jobModel.StateRunning && (job.State != jobModel.StateCanceled || resultErr != "") {
		return
	}

	now := time.Now()
	nowLocal := datetime.Local(now)

	job.EditedDate = &nowLocal
	job.Result = resultData
	job.RunBy = ""

	if resultErr != "" {
		job.ConsecutiveFailures++
		job.State = jobModel.StateFailed
		job.LastFailureError = resultErr
	} else {
		job.State = jobModel.StateFinished
		job.RunProgress = ProgressMax
	}

	// Reschedule recurring jobs
	//nolint:nestif // refactor if needed
	if job.IsRecurring {
		if resultErr == "" {
			job.ConsecutiveFailures = 0
		}

		exp, err := cronexpr.Parse(job.RecurringCron)
		if err != nil {
			job.State = jobModel.StateFailed

			messages := fmt.Sprintf("failed to parse cron string: %s", err.Error())
			if job.LastFailureError != "" {
				messages = messages + "; " + job.LastFailureError
			}

			job.LastFailureError = messages
		} else {
			job.State = jobModel.StateScheduled
			nextTime := exp.Next(now)
			nextTimeLocal := datetime.Local(nextTime)
			job.Scheduled = &nextTimeLocal
		}

		return
	}

	// Reschedule the failed job if retrying is allowed
	if job.State == jobModel.StateFailed && job.ConsecutiveFailures <= job.MaxRetries {
		const retryDelay = 15 * time.Second
		job.State = jobModel.StateScheduled
		retryTime := now.Add(retryDelay)
		retryTimeLocal := datetime.Local(retryTime)
		job.Scheduled = &retryTimeLocal
		job.RunProgress = ProgressMin
	}
}

// RunRecurringJob creates or upserts a recurring job that runs according to the given cron expression.
// This is the public API for registering recurring jobs externally (e.g. mirror sync).
func (s *Scheduler) RunRecurringJob(ctx context.Context, jobName, jobType, cronDef string, data string, maxDuration time.Duration) error {
	return s.addRecurringJob(ctx, jobName, jobType, cronDef, data, maxDuration)
}

// RunJob runs a single job.
// If a job with jobName already exists and is not running, its definition is updated.
// If it is running, the call is silently ignored.
// If no such job exists, a new one is created.
func (s *Scheduler) RunJob(ctx context.Context, jobName, jobType string, data string, maxDurationSeconds int) error {
	now := time.Now()
	nowLocal := datetime.Local(now)

	existingJob, err := s.jobStore.GetByName(ctx, jobName)
	if err != nil && !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return fmt.Errorf("failed to lookup job by name: %w", err)
	}

	if existingJob != nil {
		if existingJob.State == jobModel.StateRunning {
			return nil
		}

		existingJob.Type = jobType
		existingJob.Priority = jobModel.PriorityNormal
		existingJob.Data = data
		existingJob.MaxDurationSeconds = maxDurationSeconds
		existingJob.MaxRetries = 0
		existingJob.State = jobModel.StateScheduled
		existingJob.Scheduled = &nowLocal
		existingJob.EditedDate = datetime.NowLocalTime()

		if err := s.jobStore.UpdateDefinition(ctx, existingJob); err != nil {
			return fmt.Errorf("failed to update existing job: %w", err)
		}
	} else {
		job := &jobModel.ScheduleJob{
			Name:                jobName,
			Type:                jobType,
			Priority:            jobModel.PriorityNormal,
			Data:                data,
			MaxDurationSeconds:  maxDurationSeconds,
			MaxRetries:          0,
			State:               jobModel.StateScheduled,
			Scheduled:           &nowLocal,
			TotalExecutions:     0,
			RunBy:               "",
			RunDeadline:         nil,
			RunProgress:         ProgressMin,
			LastExecuted:        nil,
			IsRecurring:         false,
			RecurringCron:       "",
			ConsecutiveFailures: 0,
			LastFailureError:    "",
			CreatedDate:         datetime.NowLocalTime(),
			EditedDate:          datetime.NowLocalTime(),
		}

		if err := s.jobStore.Create(ctx, job); err != nil {
			return fmt.Errorf("failed to create new job: %w", err)
		}
	}

	s.scheduleProcessing(now)

	return nil
}

// RunJobs runs several jobs. It's more efficient than calling RunJob several times.
func (s *Scheduler) RunJobs(ctx context.Context, groupID string, jobs []*jobModel.ScheduleJob) error {
	if len(jobs) == 0 {
		return nil
	}

	now := time.Now()
	for _, job := range jobs {
		job.GroupID = groupID
		job.CreatedDate = datetime.NowLocalTime()
		job.EditedDate = datetime.NowLocalTime()

		if err := s.jobStore.Create(ctx, job); err != nil {
			return fmt.Errorf("failed to add new job to the database: %w", err)
		}
	}

	s.scheduleProcessing(now)

	return nil
}

// RegisterHandler registers an external job handler for the given job type.
// Must be called before Run() (which calls FinishRegistration()).
func (s *Scheduler) RegisterHandler(jobType string, handler pkgJobs.Handler) error {
	return s.executor.Register(jobType, handler)
}

// registerSystemJobs registers internal system jobs (purge, overdue).
func (s *Scheduler) registerSystemJobs() error {
	// Register purge job handler
	handlerPurge := newJobPurge(s.jobStore, s.mxManager, s.retentionTime)
	if err := s.executor.Register(JobTypePurge, handlerPurge); err != nil {
		return fmt.Errorf("failed to register purge job: %w", err)
	}

	// Register overdue job handler
	handlerOverdue := newJobOverdue(s.jobStore, s.mxManager, s)
	if err := s.executor.Register(JobTypeOverdue, handlerOverdue); err != nil {
		return fmt.Errorf("failed to register overdue job: %w", err)
	}

	// Register runner offline/online update
	handlerRunnerHeartbeat := newRunnerHeartbeatJob(s.runnerStore, s.mxManager, s)
	if err := s.executor.Register(jobTypeRunnerHeartbeat, handlerRunnerHeartbeat); err != nil {
		return fmt.Errorf("failed to register runner job: %w", err)
	}

	s.log.Info("[JobScheduler] System jobs registered successfully")
	return nil
}

// createSystemJobs creates system jobs in the database if they don't exist.
func (s *Scheduler) createSystemJobs(ctx context.Context) error {
	mx, err := globalLock(ctx, s.mxManager)
	if err != nil {
		return fmt.Errorf("failed to obtain global lock: %w", err)
	}

	defer func() {
		if err := mx.Unlock(ctx); err != nil {
			s.log.Errorw("[JobScheduler] Failed to release global lock", "error", err)
		}
	}()

	// Create purge job
	if err := s.addRecurringJob(ctx, JobUIDPurge, JobTypePurge, JobCronPurge, "", 5*time.Second); err != nil {
		return fmt.Errorf("failed to create purge job: %w", err)
	}

	// Create overdue job
	if err := s.addRecurringJob(ctx, JobUIDOverdue, JobTypeOverdue, JobCronOverdue, "", 5*time.Second); err != nil {
		return fmt.Errorf("failed to create overdue job: %w", err)
	}

	// Create runner job
	if err := s.addRecurringJob(ctx, jobUIDRunnerHeartbeat, jobTypeRunnerHeartbeat, jobCronRunnerHeartbeat, "", jobMaxDurationRunnerHeartbeat); err != nil {
		return fmt.Errorf("failed to create runner job: %w", err)
	}

	s.log.Info("[JobScheduler] System jobs created successfully")
	return nil
}

// addRecurringJob adds a recurring job to the database.
func (s *Scheduler) addRecurringJob(
	ctx context.Context,
	jobUID, jobType, cronDef string,
	data string,
	maxDur time.Duration,
) error {
	// Parse cron expression to get next execution time
	cronExp, err := cronexpr.Parse(cronDef)
	if err != nil {
		return fmt.Errorf("invalid cron definition string for job type=%s: %w", jobType, err)
	}

	now := time.Now()
	nowLocal := datetime.Local(now)
	nextExec := cronExp.Next(now)
	nextExecLocal := datetime.Local(nextExec)

	job := &jobModel.ScheduleJob{
		Name:                jobUID,
		Type:                jobType,
		Priority:            jobModel.PriorityElevated,
		Data:                data,
		Result:              "",
		MaxDurationSeconds:  int(maxDur / time.Second),
		MaxRetries:          0,
		State:               jobModel.StateScheduled,
		Scheduled:           &nextExecLocal,
		TotalExecutions:     0,
		RunBy:               "",
		RunDeadline:         nil,
		RunProgress:         ProgressMin,
		LastExecuted:        nil,
		IsRecurring:         true,
		RecurringCron:       cronDef,
		ConsecutiveFailures: 0,
		LastFailureError:    "",
		CreatedDate:         &nowLocal,
		EditedDate:          &nowLocal,
	}

	if err := s.jobStore.Upsert(ctx, job); err != nil {
		return fmt.Errorf("failed to upsert job id=%s type=%s: %w", jobUID, jobType, err)
	}

	return nil
}

// GetJobProgress retrieves the progress of a job.
func (s *Scheduler) GetJobProgress(ctx context.Context, jobName string) (map[string]interface{}, error) {
	job, err := s.jobStore.GetByName(ctx, jobName)
	if err != nil {
		return nil, err
	}

	return mapToProgress(job), nil
}

// GetJobProgressForGroup retrieves the progress of jobs in a group.
func (s *Scheduler) GetJobProgressForGroup(ctx context.Context, jobGroupID string) ([]map[string]interface{}, error) {
	jobs, err := s.jobStore.ListByGroupID(ctx, jobGroupID)
	if err != nil {
		return nil, err
	}
	return mapToProgressMany(jobs), nil
}

// PurgeJobsByGroupID deletes all jobs for a group id.
func (s *Scheduler) PurgeJobsByGroupID(ctx context.Context, jobGroupID string) (int64, error) {
	n, err := s.jobStore.DeleteByGroupID(ctx, jobGroupID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete jobs by group id=%s: %w", jobGroupID, err)
	}
	return n, nil
}

// publishStateChange publishes a job state change message to the pubsub service.
// 发布作业状态变更消息到 pubsub 服务
func publishStateChange(ctx context.Context, ps pubsub.PubSub, job *jobModel.ScheduleJob) error {
	payload := map[string]interface{}{
		"name":     job.Name,
		"type":     job.Type,
		"state":    job.State,
		"progress": job.RunProgress,
		"result":   job.Result,
		"failure":  job.LastFailureError,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job state change: %w", err)
	}

	return ps.Publish(ctx, PubSubTopicStateChange, data)
}

func mapToProgressMany(jobs []*jobModel.ScheduleJob) []map[string]interface{} {
	if jobs == nil {
		return nil
	}
	j := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		j[i] = mapToProgress(job)
	}
	return j
}

func mapToProgress(job *jobModel.ScheduleJob) map[string]interface{} {
	return map[string]interface{}{
		"state":        job.State,
		"progress":     job.RunProgress,
		"result":       job.Result,
		"failure":      job.LastFailureError,
		"lastExecuted": job.LastExecuted,
	}
}
