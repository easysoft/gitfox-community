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
	log.Info().Msgf("[测试]: worker=%s ready to exe stage", w.token)
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
		log.Info().Msgf("[测试]: worker=%s received stageID=%d", w.token, b.ID)
		return b, nil
	}
}

//nolint:gocognit // refactor if needed.
func (q *queue) signal(ctx context.Context) error {
	log.Info().Msg("[测试]: queue signal triggered")
	if err := q.globMx.Lock(ctx); err != nil {
		log.Info().Err(err).Msg("[测试]: failed to acquire global lock")
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
	log.Info().Msgf("[测试]: queue state snapshot, workers=%d, paused=%v", count, pause)
	if pause {
		log.Info().Msg("[测试]: queue is paused, skipping signal")
		return nil
	}
	if count == 0 {
		log.Info().Msg("[测试]: no workers available, skipping signal")
		return nil
	}
	items, err := q.store.ListIncomplete(ctx)
	if err != nil {
		log.Info().Err(err).Msg("[测试]: failed to list incomplete stages")
		return err
	}
	log.Info().Msgf("[测试]: incomplete stages fetched from DB, count=%d", len(items))

	// 查找数据库 runners 表中所有未删除，未停用，在线的 agent runner
	availableRunners, err := q.getAvailableRunners(ctx)
	if err != nil {
		log.Info().Err(err).Msg("[测试]: failed to get available runners")
		return err
	}

	// 初始化可用 runner token 列表
	var availableRunnerTokenList []string
	if len(availableRunners) != 0 {
		for _, runner := range availableRunners {
			if runner.Token != "" {
				availableRunnerTokenList = append(availableRunnerTokenList, runner.Token)
			}
		}
	}
	log.Info().Msgf("[测试]: available agent runner tokens, count=%d", len(availableRunnerTokenList))

	q.Lock()
	defer q.Unlock()
	log.Info().Msgf("[测试]: starting stage-worker matching, stages=%d, workers=%d", len(items), len(q.workers))
	for _, item := range items {
		if item.Status == constant.CIStatusRunning {
			log.Info().Msgf("[测试]: skip stage (already running), stage_id=%d", item.ID)
			continue
		}
		log.Info().Msgf("[测试]: trying to match stage, stage_id=%d, type=%s, os=%s, arch=%s", item.ID, item.Type, item.OS, item.Arch)

		matched := false
	loop:
		for w := range q.workers {
			// the worker must match the resource kind and type
			if !matchResource(w.typ, item.Type) {
				log.Info().Msgf("[测试]: worker skipped (type mismatch), stage_id=%d, worker_type=%s, stage_type=%s", item.ID, w.typ, item.Type)
				continue
			}

			// agent 类型的 runner 增加校验逻辑
			if w.typ == "agent" {
				if len(availableRunnerTokenList) == 0 {
					log.Info().Msgf("[测试]: worker skipped (agent type but no available runner tokens), stage_id=%d", item.ID)
					continue
				}
				if !slices.Contains(availableRunnerTokenList, w.token) {
					log.Info().Msgf("[测试]: worker skipped (agent token not in available list), stage_id=%d, worker_token=%s", item.ID, w.token)
					continue
				}
			}

			if w.os != "" || w.arch != "" {
				if w.os != item.OS {
					log.Info().Msgf("[测试]: worker skipped (OS mismatch), stage_id=%d, worker_os=%s, stage_os=%s", item.ID, w.os, item.OS)
					continue
				}
				if w.arch != item.Arch {
					log.Info().Msgf("[测试]: worker skipped (arch mismatch), stage_id=%d, worker_arch=%s, stage_arch=%s", item.ID, w.arch, item.Arch)
					continue
				}
			}

			if len(item.Labels) > 0 || len(w.labels) > 0 {
				if !checkLabels(item.Labels, w.labels) {
					log.Info().Msgf("[测试]: worker skipped (labels mismatch), stage_id=%d", item.ID)
					continue
				}
			}

			log.Info().Msgf("[测试]: worker matched, dispatching stage, worker_token=%s, stage_id=%d, type=%s", w.token, item.ID, w.typ)

			select {
			case w.channel <- item:
				log.Info().Msgf("[测试]: stage dispatched to worker successfully, worker_token=%s, stage_id=%d", w.token, item.ID)
			case <-w.done:
				log.Info().Msgf("[测试]: worker cancelled before dispatch, worker_token=%s, stage_id=%d", w.token, item.ID)
			}

			delete(q.workers, w)
			matched = true
			break loop
		}
		if !matched {
			log.Info().Msgf("[测试]: no matching worker found for stage, stage_id=%d", item.ID)
		}
	}
	log.Info().Msgf("[测试]: signal completed, remaining_workers=%d", len(q.workers))
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
