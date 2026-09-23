// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package queue

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// 提供了一个[Runner]，当收到消息时，它可以按名称运行已注册的job [Func]s
// 底层队列。

// -限制可同时运行的作业数量
// -作业运行时自动消息超时扩展
// -安全关机

func NewRunner(opts NewRunnerOpts) *Runner {
	if opts.Log == nil {
		opts.Log = &discardLogger{}
	}

	if opts.Limit == 0 {
		opts.Limit = runtime.GOMAXPROCS(0)
	}

	if opts.PollInterval == 0 {
		opts.PollInterval = 100 * time.Millisecond
	}

	if opts.Extend == 0 {
		opts.Extend = 5 * time.Second
	}

	return &Runner{
		extend:        opts.Extend,
		jobCountLimit: opts.Limit,
		jobs:          make(map[string]Func),
		log:           opts.Log,
		pollInterval:  opts.PollInterval,
		queue:         opts.Queue,
	}
}

// 启动Runner，直到给定的上下文被取消。
// 当上下文被取消时，等待作业完成。
func (r *Runner) Start(ctx context.Context) {
	var names []string
	for k := range r.jobs {
		names = append(names, k)
	}
	sort.Strings(names)

	r.log.Info("Starting job runner", "jobs", names)

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			r.log.Info("Stopping job runner")
			wg.Wait()
			r.log.Info("Stopped job runner")
			return
		default:
			r.receiveAndRun(ctx, &wg)
		}
	}
}

func (r *Runner) receiveAndRun(ctx context.Context, wg *sync.WaitGroup) {
	r.jobCountLock.RLock()
	if r.jobCount == r.jobCountLimit {
		r.jobCountLock.RUnlock()
		// This is to avoid a busy loop
		time.Sleep(r.pollInterval)
		return
	} else {
		r.jobCountLock.RUnlock()
	}

	m, err := r.queue.ReceiveAndWait(ctx, r.pollInterval)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return
		}
		r.log.Info("Error receiving job", "error", err)
		// Sleep a bit to not hammer the queue if there's an error with it
		time.Sleep(time.Second)
		return
	}

	if m == nil {
		return
	}

	var jm message
	if err := gob.NewDecoder(bytes.NewReader(m.Body)).Decode(&jm); err != nil {
		r.log.Info("Error decoding job message body", "error", err)
		return
	}

	job, ok := r.jobs[jm.Name]
	if !ok {
		panic(fmt.Sprintf(`job "%v" not registered`, jm.Name))
	}

	r.jobCountLock.Lock()
	r.jobCount++
	r.jobCountLock.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			r.jobCountLock.Lock()
			r.jobCount--
			r.jobCountLock.Unlock()
		}()

		defer func() {
			if rec := recover(); rec != nil {
				r.log.Info("Recovered from panic in job", "name", jm.Name, "id", m.ID, "error", rec)
			}
		}()

		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Extend the job message while the job is running
		go func() {
			// Start by sleeping so we don't extend immediately
			time.Sleep(r.extend - r.extend/5)
			for {
				select {
				case <-jobCtx.Done():
					return
				default:
					r.log.Info("Extending message timeout", "name", jm.Name, "id", m.ID)
					if err := r.queue.Extend(jobCtx, m.ID, r.extend); err != nil {
						r.log.Info("Error extending message timeout", "name", jm.Name, "id", m.ID, "error", err)
					}
					time.Sleep(r.extend - r.extend/5)
				}
			}
		}()

		r.log.Info("Running job", "name", jm.Name, "id", m.ID)
		before := time.Now()
		if err := job(jobCtx, jm.Message); err != nil {
			r.log.Info("Error running job", "name", jm.Name, "id", m.ID, "error", err)
			return
		}
		duration := time.Since(before)
		r.log.Info("Ran job", "name", jm.Name, "id", m.ID, "duration", duration)

		deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		if err := r.queue.Delete(deleteCtx, m.ID); err != nil {
			r.log.Info("Error deleting job from queue, it will be retried", "name", jm.Name, "id", m.ID, "error", err)
		}
	}()
}

// Func 是一项待完成的任务。它从队列中获取消息
type Func func(ctx context.Context, m []byte) error

func (r *Runner) Register(name string, job Func) {
	if _, ok := r.jobs[name]; ok {
		panic(fmt.Sprintf(`job "%v" already registered`, name))
	}
	r.jobs[name] = job
}

// Create 为给定队列中的指定作业创建消息。
func Create(ctx context.Context, q *Queue, name string, m Message) (ID, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(message{Name: name, Message: m.Body}); err != nil {
		return "", err
	}
	return q.SendAndGetID(ctx, Message{Body: buf.Bytes(), Delay: m.Delay, Priority: m.Priority})
}

// CreateTx 类似于 Create，但在现有事务中。
func CreateTx(ctx context.Context, tx *sql.Tx, q *Queue, name string, m Message) (ID, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(message{Name: name, Message: m.Body}); err != nil {
		return "", err
	}
	return q.SendAndGetIDTx(ctx, tx, Message{Body: buf.Bytes(), Delay: m.Delay, Priority: m.Priority})
}

// logger 是一个日志记录器，它匹配 slog.Logger 的 info 级别方法。
type logger interface {
	Info(msg string, args ...any)
}

type discardLogger struct{}

func (d *discardLogger) Info(msg string, args ...any) {}
