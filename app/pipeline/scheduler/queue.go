package scheduler

import (
	"context"
	"slices"
	"sync"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	runnerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/runner"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"github.com/rs/zerolog/log"
)

type queue struct {
	sync.Mutex
	globMx lock.Mutex

	ready       chan struct{}
	paused      bool
	interval    time.Duration
	store       stageStore.Interface
	runnerStore runnerStore.Interface
	workers     map[*worker]struct{}
	ctx         context.Context
}

// newQueue returns a new Queue backed by the build datastore.
func newQueue(store stageStore.Interface, runnerStore runnerStore.Interface, lock lock.MutexManager) (*queue, error) {

	const lockKey = "build_queue"
	mx, err := lock.NewMutex(lockKey)
	if err != nil {
		return nil, err
	}
	q := &queue{
		store:       store,
		runnerStore: runnerStore,
		globMx:      mx,
		ready:       make(chan struct{}, 1),
		workers:     map[*worker]struct{}{},
		interval:    time.Minute,
		ctx:         context.Background(),
	}
	go func() {
		if err := q.start(); err != nil {
			log.Err(err).Msg("queue start failed")
		}
	}()

	return q, nil
}

// getAvailableRunners 获取数据库中所有未删除，未停用，在线的 runner
func (q *queue) getAvailableRunners(ctx context.Context) ([]*runnerModel.Runner, error) {
	filter := &runnerModel.ReqFilter{
		Page:     1,
		PageSize: 10000,
		Online:   "online", // 只查询在线状态的 runner
		Status:   "enable", // 只查询启用状态的 runner
	}

	res, err := q.runnerStore.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Debug().
		Int("total_runners_found", len(res.Data)).
		Msg("retrieved available runners from database")

	return res.Data, nil
}

func (q *queue) Schedule() error {
	select {
	case q.ready <- struct{}{}:
	default:
	}
	return nil
}

func (q *queue) Pause(_ context.Context) error {
	q.Lock()
	q.paused = true
	q.Unlock()
	return nil
}

func (q *queue) Request(ctx context.Context, params Filter) (*stageModel.Stage, error) {
	w := &worker{
		token:    params.Token,
		typ:      params.Type,
		os:       params.OS,
		arch:     params.Arch,
		labels:   params.Labels,
		metadata: params.Metadata,
		channel:  make(chan *stageModel.Stage),
		done:     ctx.Done(),
	}
	log.Ctx(ctx).Debug().Msgf("[queue]: add new worker %+v", w)
	q.Lock()
	q.workers[w] = struct{}{}
	q.Unlock()

	select {
	case q.ready <- struct{}{}:
	default:
	}

	select {
	case <-ctx.Done():
		q.Lock()
		delete(q.workers, w)
		q.Unlock()
		return nil, ctx.Err()
	case b := <-w.channel:
		return b, nil
	}
}

//nolint:gocognit // refactor if needed.
func (q *queue) signal(ctx context.Context) error {
	if err := q.globMx.Lock(ctx); err != nil {
		return err
	}
	defer func() {
		if err := q.globMx.Unlock(ctx); err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to release global lock after signaling")
		}
	}()

	q.Lock()
	count := len(q.workers)
	pause := q.paused
	q.Unlock()
	if pause {
		return nil
	}
	if count == 0 {
		return nil
	}
	items, err := q.store.ListIncomplete(ctx)
	if err != nil {
		return err
	}
	log.Ctx(ctx).Info().Msgf("skip signal for now: %d incomplete stages", len(items))
	items = []*stageModel.Stage{} // TODO: remove this line after testing

	// 查找数据库 runners 表中所有未删除，未停用，在线的 runner
	availableRunners, err := q.getAvailableRunners(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get available runners")
		return err
	}

	// 初始化可用 runner token 列表
	var availableRunnerTokenList []string
	if len(availableRunners) != 0 {
		// 遍历可用 runners，提取每个 runner 的 token 并拼装至 availableRunnerTokenList
		for _, runner := range availableRunners {
			if runner.Token != "" {
				availableRunnerTokenList = append(availableRunnerTokenList, runner.Token)
			}
		}
	}

	q.Lock()
	defer q.Unlock()
	for _, item := range items {
		if item.Status == constant.CIStatusRunning {
			continue
		}

	loop:
		for w := range q.workers {
			// the worker must match the resource kind and type
			if !matchResource(w.typ, item.Type) {
				continue
			}

			// agent 类型的 runner 增加校验逻辑
			if w.typ == "agent" {
				if len(availableRunnerTokenList) == 0 {
					continue
				}
				if !slices.Contains(availableRunnerTokenList, w.token) {
					continue
				}
			}

			if w.os != "" || w.arch != "" {
				// the worker is platform-specific. check to ensure
				// the queue item matches the worker platform.
				if w.os != item.OS {
					continue
				}
				if w.arch != item.Arch {
					continue
				}
			}

			if len(item.Labels) > 0 || len(w.labels) > 0 {
				if !checkLabels(item.Labels, w.labels) {
					continue
				}
			}

			log.Info().Msgf("Runner[%s] stage[%d] selected to exec stage", w.token, item.ID)

			select {
			case w.channel <- item:
			case <-w.done:
			}

			delete(q.workers, w)
			break loop
		}
	}
	return nil
}

func (q *queue) start() error {
	for {
		select {
		case <-q.ctx.Done():
			return q.ctx.Err()
		case <-q.ready:
			if err := q.signal(q.ctx); err != nil {
				// don't return, only log error
				log.Ctx(q.ctx).Err(err).Msg("failed to signal on ready")
			}
		case <-time.After(q.interval):
			if err := q.signal(q.ctx); err != nil {
				// don't return, only log error
				log.Ctx(q.ctx).Err(err).Msg("failed to signal on interval")
			}
		}
	}
}

type worker struct {
	token    string
	typ      string
	os       string
	arch     string
	labels   map[string]string
	metadata map[string]string
	channel  chan *stageModel.Stage
	done     <-chan struct{}
}

func checkLabels(a, b map[string]string) bool {
	//if len(a) != len(b) {
	//	return false
	//}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

// matchResource is a helper function that returns.
func matchResource(typea, typeb string) bool {
	if typea == "" {
		typea = "docker"
	}
	if typeb == "" {
		typeb = "docker"
	}
	return typea == typeb
}
