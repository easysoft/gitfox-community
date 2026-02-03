package jobs

import (
	"context"
	"runtime/debug"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

// Executor holds map of Handler objects per each job type registered.
// The Scheduler uses the Executor to start execution of jobs.
// 执行器 负责执行作业 并将结果发布到事件总线
type Executor struct {
	handlerMap      map[string]Handler
	handlerComplete bool
	publisher       pubsub.Publisher
}

// ProgressReporter can be used by a job Handler to report back the execution progress.
// 进度报告器 可以由作业处理程序使用 来报告执行进度
type ProgressReporter func(progress int, result string) ([]byte, error)

// Handler is a job executor for a specific job type.
// An implementation should try to honor the context and
// try to abort the execution as soon as the context is done.
// Handler是特定作业类型的作业执行器。
// 一旦上下文完成，尝试中止执行。
type Handler interface {
	Handle(ctx context.Context, input string, reportFn ProgressReporter) (result string, err error)
}

var errNoHandlerDefined = errors.New("no handler registered for the job type")

// NewExecutor creates new Executor.
func NewExecutor(publisher pubsub.Publisher) *Executor {
	return &Executor{
		handlerMap:      make(map[string]Handler),
		handlerComplete: false,
		publisher:       publisher,
	}
}

// Register registers a job Handler for the provided job type.
// This function is not thread safe. All calls are expected to be made
// in a single thread during the application boot time.
func (e *Executor) Register(jobType string, exec Handler) error {
	if jobType == "" {
		return errors.New("jobType must not be empty")
	}

	if e.handlerComplete {
		return errors.New("job handler registration is complete")
	}

	if exec == nil {
		return errors.New("provided Handler is nil")
	}

	if _, ok := e.handlerMap[jobType]; ok {
		return errors.Newf("a Handler is already defined to run the '%s' job types", jobType)
	}

	e.handlerMap[jobType] = exec

	return nil
}

// FinishRegistration forbids further registration of job types.
// It is called by the Scheduler when it starts.
// 完成注册 禁止进一步注册作业类型
func (e *Executor) FinishRegistration() {
	e.handlerComplete = true
}

// exec runs a single job. This function is synchronous,
// so the caller is responsible to run it in a separate go-routine.
// 执行一个作业 这个函数是同步的 所以调用者需要在单独的 goroutine 中运行它
func (e *Executor) Exec(ctx context.Context, jobName, jobType string, input string, reportFn ProgressReporter) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Newf(
				"panic while processing job=%s type=%s: %v\n%s",
				jobName, jobType, r, debug.Stack())
		}
	}()

	exec, ok := e.handlerMap[jobType]
	if !ok {
		return "", errNoHandlerDefined
	}

	return exec.Handle(ctx, input, reportFn) // runs the job
}
