package scheduler

// import (
// 	"context"
// 	"slices"
// 	"sync"
// 	"time"

// 	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
// 	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
// 	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
// 	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
// 	"github.com/easysoft/gitfox/types/enum"
// 	"github.com/rs/zerolog/log"
// )

// type queue struct {
// 	sync.Mutex
// 	globMx lock.Mutex

// 	ready       chan struct{}
// 	paused      bool
// 	interval    time.Duration
// 	store       stageStore.Interface
// 	runnerStore store.RunnerStore
// 	workers     map[*worker]struct{}
// 	ctx         context.Context
// }

// // newQueue returns a new Queue backed by the build datastore.
// func newQueue(store stageStore.Interface, lock lock.MutexManager) (*queue, error) {

// 	const lockKey = "build_queue"
// 	mx, err := lock.NewMutex(lockKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	q := &queue{
// 		store:       store,
// 		runnerStore: runnerStore,
// 		globMx:      mx,
// 		ready:       make(chan struct{}, 1),
// 		workers:     map[*worker]struct{}{},
// 		interval:    time.Minute,
// 		ctx:         context.Background(),
// 	}
// 	go func() {
// 		if err := q.start(); err != nil {
// 			log.Err(err).Msg("queue start failed")
// 		}
// 	}()

// 	return q, nil
// }

// // getAvailableRunners 获取数据库中所有未删除，未停用，在线的 runner
// func (q *queue) getAvailableRunners(ctx context.Context) ([]types.RunnerInfo, error) {
// 	filter := &types.RunnerFilter{
// 		ListQueryFilter: types.ListQueryFilter{
// 			Pagination: types.Pagination{
// 				Size: 1000, // 设置足够大的数量以获取所有记录
// 			},
// 		},
// 		Online: "online", // 只查询在线状态的 runner
// 		Status: "enable", // 只查询启用状态的 runner
// 	}

// 	runners, _, err := q.runnerStore.List(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Ctx(ctx).Debug().
// 		Int("total_runners_found", len(runners)).
// 		Msg("retrieved available runners from database")

// 	return runners, nil
// }

// func (q *queue) Schedule() error {
// 	select {
// 	case q.ready <- struct{}{}:
// 	default:
// 	}
// 	return nil
// }

// func (q *queue) Pause(_ context.Context) error {
// 	q.Lock()
// 	q.paused = true
// 	q.Unlock()
// 	return nil
// }

// func (q *queue) Request(ctx context.Context, params Filter) (*stageModel.Stage, error) {
// 	w := &worker{
// 		token:    params.Token,
// 		kind:     params.Kind,
// 		typ:      params.Type,
// 		os:       params.OS,
// 		arch:     params.Arch,
// 		kernel:   params.Kernel,
// 		variant:  params.Variant,
// 		labels:   params.Labels,
// 		metadata: params.Metadata,
// 		channel:  make(chan *stageModel.Stage),
// 		done:     ctx.Done(),
// 	}
// 	log.Ctx(ctx).Debug().Msgf("[queue]: add new worker %+v", w)
// 	q.Lock()
// 	q.workers[w] = struct{}{}
// 	updateRunnerHeartBeat(w)
// 	q.Unlock()

// 	select {
// 	case q.ready <- struct{}{}:
// 	default:
// 	}

// 	select {
// 	case <-ctx.Done():
// 		q.Lock()
// 		delete(q.workers, w)
// 		q.Unlock()
// 		return nil, ctx.Err()
// 	case b := <-w.channel:
// 		return b, nil
// 	}
// }

// //nolint:gocognit // refactor if needed.
// func (q *queue) signal(ctx context.Context) error {
// 	if err := q.globMx.Lock(ctx); err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if err := q.globMx.Unlock(ctx); err != nil {
// 			log.Ctx(ctx).Err(err).Msg("failed to release global lock after signaling")
// 		}
// 	}()

// 	q.Lock()
// 	count := len(q.workers)
// 	pause := q.paused
// 	q.Unlock()
// 	if pause {
// 		return nil
// 	}
// 	if count == 0 {
// 		return nil
// 	}
// 	items, err := q.store.ListIncomplete(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// 查找数据库 runners 表中所有未删除，未停用，在线的 runner
// 	availableRunners, err := q.getAvailableRunners(ctx)
// 	if err != nil {
// 		log.Ctx(ctx).Error().Err(err).Msg("failed to get available runners")
// 		return err
// 	}

// 	// 初始化可用 runner token 列表
// 	var availableRunnerTokenList []string
// 	if len(availableRunners) != 0 {
// 		// 遍历可用 runners，提取每个 runner 的 token 并拼装至 availableRunnerTokenList
// 		for _, runner := range availableRunners {
// 			if runner.Token != "" {
// 				availableRunnerTokenList = append(availableRunnerTokenList, runner.Token)
// 			}
// 		}
// 	}

// 	q.Lock()
// 	defer q.Unlock()
// 	for _, item := range items {
// 		if item.Status == constant.CIStatusRunning {
// 			continue
// 		}

// 		// if the stage defines concurrency limits we
// 		// need to make sure those limits are not exceeded
// 		// before proceeding.
// 		if !withinLimits(item, items) {
// 			continue
// 		}

// 		// if the system defines concurrency limits
// 		// per repository we need to make sure those limits
// 		// are not exceeded before proceeding.
// 		if shouldThrottle(item, items, item.LimitRepo) {
// 			continue
// 		}

// 	loop:
// 		for w := range q.workers {
// 			// the worker must match the resource kind and type
// 			if !matchResource(w.kind, w.typ, item.Kind, item.Type) {
// 				continue
// 			}

// 			// agent 类型的 runner 增加校验逻辑
// 			if w.typ == "exec" && w.kind == "pipeline" {
// 				if len(availableRunnerTokenList) == 0 {
// 					continue
// 				}
// 				if !slices.Contains(availableRunnerTokenList, w.token) {
// 					continue
// 				}
// 			}

// 			if w.os != "" || w.arch != "" || w.variant != "" || w.kernel != "" {
// 				// the worker is platform-specific. check to ensure
// 				// the queue item matches the worker platform.
// 				if w.os != item.OS {
// 					continue
// 				}
// 				if w.arch != item.Arch {
// 					continue
// 				}
// 				// if the pipeline defines a variant it must match
// 				// the worker variant (e.g. arm6, arm7, etc).
// 				if item.Variant != "" && item.Variant != w.variant {
// 					continue
// 				}
// 				// if the pipeline defines a kernel version it must match
// 				// the worker kernel version (e.g. 1709, 1803).
// 				if item.Kernel != "" && item.Kernel != w.kernel {
// 					continue
// 				}
// 			}

// 			if len(item.Labels) > 0 || len(w.labels) > 0 {
// 				if !checkLabels(item.Labels, w.labels) {
// 					continue
// 				}
// 			}

// 			log.Info().Msgf("Runner[%s] stage[%d] selected to exec stage", w.token, item.ID)

// 			select {
// 			case w.channel <- item:
// 			case <-w.done:
// 			}

// 			delete(q.workers, w)
// 			break loop
// 		}
// 	}
// 	return nil
// }

// func (q *queue) start() error {
// 	for {
// 		select {
// 		case <-q.ctx.Done():
// 			return q.ctx.Err()
// 		case <-q.ready:
// 			if err := q.signal(q.ctx); err != nil {
// 				// don't return, only log error
// 				log.Ctx(q.ctx).Err(err).Msg("failed to signal on ready")
// 			}
// 		case <-time.After(q.interval):
// 			if err := q.signal(q.ctx); err != nil {
// 				// don't return, only log error
// 				log.Ctx(q.ctx).Err(err).Msg("failed to signal on interval")
// 			}
// 		}
// 	}
// }

// type worker struct {
// 	token    string
// 	kind     string
// 	typ      string
// 	os       string
// 	arch     string
// 	kernel   string
// 	variant  string
// 	labels   map[string]string
// 	metadata map[string]string
// 	channel  chan *types.Stage
// 	done     <-chan struct{}
// }

// func checkLabels(a, b map[string]string) bool {
// 	//if len(a) != len(b) {
// 	//	return false
// 	//}
// 	for k, v := range a {
// 		if w, ok := b[k]; !ok || v != w {
// 			return false
// 		}
// 	}
// 	return true
// }

// func withinLimits(stage *types.Stage, siblings []*types.Stage) bool {
// 	if stage.Limit == 0 {
// 		return true
// 	}
// 	count := 0
// 	for _, sibling := range siblings {
// 		if sibling.RepoID != stage.RepoID {
// 			continue
// 		}
// 		if sibling.ID == stage.ID {
// 			continue
// 		}
// 		if sibling.Name != stage.Name {
// 			continue
// 		}
// 		if sibling.ID < stage.ID ||
// 			sibling.Status == enum.CIStatusRunning {
// 			count++
// 		}
// 	}
// 	return count < stage.Limit
// }

// func shouldThrottle(stage *types.Stage, siblings []*types.Stage, limit int) bool {
// 	// if no throttle limit is defined (default) then
// 	// return false to indicate no throttling is needed.
// 	if limit == 0 {
// 		return false
// 	}
// 	// if the repository is running it is too late
// 	// to skip and we can exit
// 	if stage.Status == enum.CIStatusRunning {
// 		return false
// 	}

// 	count := 0
// 	// loop through running stages to count number of
// 	// running stages for the parent repository.
// 	for _, sibling := range siblings {
// 		// ignore stages from other repository.
// 		if sibling.RepoID != stage.RepoID {
// 			continue
// 		}
// 		// ignore this stage and stages that were
// 		// scheduled after this stage.
// 		if sibling.ID >= stage.ID {
// 			continue
// 		}
// 		count++
// 	}
// 	// if the count of running stages exceeds the
// 	// throttle limit return true.
// 	return count >= limit
// }

// // matchResource is a helper function that returns.
// func matchResource(kinda, typea, kindb, typeb string) bool {
// 	if kinda == "" {
// 		kinda = "pipeline"
// 	}
// 	if kindb == "" {
// 		kindb = "pipeline"
// 	}
// 	if typea == "" {
// 		typea = "docker"
// 	}
// 	if typeb == "" {
// 		typeb = "docker"
// 	}
// 	return kinda == kindb && typea == typeb
// }
