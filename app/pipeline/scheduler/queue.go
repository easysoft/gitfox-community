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
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
			logger.GetLogger(q.ctx).Errorf("queue start failed, err=%v", err)
		}
	}()

	return q, nil
}

// getAvailableRunners 获取数据库中所有未删除，未停用，在线的 runner（排除内置 runner）
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

	// 过滤掉内置 runner（isDefault=true），内置 runner 由嵌入式 poller 管理，不需要参与 agent 调度
	var agentRunners []*runnerModel.Runner
	for _, r := range res.Data {
		if !r.IsDefault {
			agentRunners = append(agentRunners, r)
		}
	}

	logger.GetLogger(ctx).Debugf("retrieved available runners from database, total_runners_found=%d, agent_runners=%d", len(res.Data), len(agentRunners))

	return agentRunners, nil
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
	logger := logger.GetLogger(ctx)
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
	logger.Infof("[pipeline]: worker=%s ready to exe stage", w.token)
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
		logger.Infof("[pipeline]: worker=%s received stageID=%d", w.token, b.ID)
		return b, nil
	}
}

//nolint:gocognit // refactor if needed.
func (q *queue) signal(ctx context.Context) error {
	logger := logger.GetLogger(ctx)
	logger.Debugf("[pipeline]: queue signal triggered")
	if err := q.globMx.Lock(ctx); err != nil {
		logger.Infof("[pipeline]: failed to acquire global lock, err=%v", err)
		return err
	}
	defer func() {
		if err := q.globMx.Unlock(ctx); err != nil {
			logger.Errorf("failed to release global lock after signaling, err=%v", err)
		}
	}()

	q.Lock()
	count := len(q.workers)
	pause := q.paused
	q.Unlock()
	logger.Debugf("[pipeline]: queue state snapshot, workers=%d, paused=%v", count, pause)
	if pause {
		logger.Debugf("[pipeline]: queue is paused, skipping signal")
		return nil
	}
	if count == 0 {
		logger.Debugf("[pipeline]: no workers available, skipping signal")
		return nil
	}
	items, err := q.store.ListIncomplete(ctx)
	if err != nil {
		logger.Infof("[pipeline]: failed to list incomplete stages, err=%v", err)
		return err
	}
	logger.Debugf("[pipeline]: incomplete stages fetched from DB, count=%d", len(items))

	// 查找数据库 runners 表中所有未删除，未停用，在线的 agent runner
	availableRunners, err := q.getAvailableRunners(ctx)
	if err != nil {
		logger.Infof("[pipeline]: failed to get available runners, err=%v", err)
		return err
	}

	// 初始化可用 runner token 列表及 token→runner 映射
	availableRunnerMap := make(map[string]*runnerModel.Runner, len(availableRunners))
	var availableRunnerTokenList []string
	for _, runner := range availableRunners {
		if runner.Token != "" {
			availableRunnerTokenList = append(availableRunnerTokenList, runner.Token)
			availableRunnerMap[runner.Token] = runner
		}
	}
	logger.Debugf("[pipeline]: available agent runner tokens, count=%d", len(availableRunnerTokenList))

	q.Lock()
	defer q.Unlock()
	logger.Debugf("[pipeline]: starting stage-worker matching, stages=%d, workers=%d", len(items), len(q.workers))
	for _, item := range items {
		if item.Status == constant.CIStatusRunning {
			logger.Infof("[pipeline]: skip stage (already running), stage_id=%d", item.ID)
			continue
		}
		if item.Host != "" {
			logger.Infof("[pipeline]: skip stage (already assigned), stage_id=%d, host=%s", item.ID, item.Host)
			continue
		}
		logger.Infof("[pipeline]: trying to match stage, stage_id=%d, type=%s, os=%s, arch=%s", item.ID, item.Type, item.OS, item.Arch)

		matched := false
	loop:
		for w := range q.workers {
			// the worker must match the resource kind and type
			if !matchResource(w.typ, item.Type) {
				logger.Debugf("[pipeline]: worker skipped (type mismatch), stage_id=%d, worker_type=%s, stage_type=%s", item.ID, w.typ, item.Type)
				continue
			}

			// agent 类型的 runner 增加校验逻辑
			if w.typ == "agent" {
				if len(availableRunnerTokenList) == 0 {
					logger.Debugf("[pipeline]: worker skipped (agent type but no available runner tokens), stage_id=%d", item.ID)
					continue
				}
				if !slices.Contains(availableRunnerTokenList, w.token) {
					logger.Debugf("[pipeline]: worker skipped (agent token not in available list), stage_id=%d, worker_token=%s", item.ID, w.token)
					continue
				}

				if w.os != "" || w.arch != "" {
					if w.os != item.OS {
						logger.Debugf("[pipeline]: worker skipped (OS mismatch), stage_id=%d, worker_os=%s, stage_os=%s", item.ID, w.os, item.OS)
						continue
					}
					if w.arch != item.Arch {
						logger.Debugf("[pipeline]: worker skipped (arch mismatch), stage_id=%d, worker_arch=%s, stage_arch=%s", item.ID, w.arch, item.Arch)
						continue
					}
				}
				// agent 用 DB 中的 labels 做匹配
				var workerLabels []string
				if runner, ok := availableRunnerMap[w.token]; ok {
					workerLabels = runner.Labels
				}
				if len(item.Labels) > 0 {
					if !checkLabels(mapKeys(item.Labels), workerLabels) {
						logger.Debugf("[pipeline]: worker skipped (labels mismatch), stage_id=%d", item.ID)
						continue
					}
				}
			}

			logger.Infof("[pipeline]: worker matched, dispatching stage, worker_token=%s, stage_id=%d, type=%s", w.token, item.ID, w.typ)

			select {
			case w.channel <- item:
				logger.Infof("[pipeline]: stage dispatched to worker successfully, worker_token=%s, stage_id=%d", w.token, item.ID)
			case <-w.done:
				logger.Infof("[pipeline]: worker cancelled before dispatch, worker_token=%s, stage_id=%d", w.token, item.ID)
			}

			delete(q.workers, w)
			matched = true
			break loop
		}
		if !matched {
			logger.Infof("[pipeline]: no matching worker found for stage, stage_id=%d", item.ID)
		}
	}
	logger.Debugf("[pipeline]: signal completed, remaining_workers=%d", len(q.workers))
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
				logger.GetLogger(q.ctx).Errorf("failed to signal on ready, err=%v", err)
			}
		case <-time.After(q.interval):
			if err := q.signal(q.ctx); err != nil {
				// don't return, only log error
				logger.GetLogger(q.ctx).Errorf("failed to signal on interval, err=%v", err)
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

// checkLabels 检查 stage 的所有 label 是否都在 runner labels 中存在
// stageLabels 和 runnerLabels 都是 []string
func checkLabels(stageLabels, runnerLabels []string) bool {
	if len(stageLabels) == 0 {
		return true
	}
	runnerSet := make(map[string]struct{}, len(runnerLabels))
	for _, l := range runnerLabels {
		runnerSet[l] = struct{}{}
	}
	for _, l := range stageLabels {
		if _, ok := runnerSet[l]; !ok {
			return false
		}
	}
	return true
}

// mapKeys 将 map[string]string 的 key 提取为 []string
// item.Labels 改为 []string 后可删除此函数
func mapKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func matchResource(typea, typeb string) bool {
	if typea == "" {
		typea = "inner"
	}
	if typeb == "" {
		typeb = "inner"
	}
	return typea == typeb
}
