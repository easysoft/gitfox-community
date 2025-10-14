package queue

import (
	"sync"
	"time"
)

// NewRunnerOpts是 [NewRunner] 的选项。

// - [NewRunner.Extend] 是在作业运行时每次延长作业消息超时的量。
// - [NewRunnerOpts.Limit] 表示可以同时运行多少个作业。
// - [NewRunner.PollInterval] 是runner在队列中轮询新消息的频率。
type NewRunnerOpts struct {
	Extend       time.Duration
	Limit        int
	Log          logger
	PollInterval time.Duration
	Queue        *Queue
}

type Runner struct {
	extend        time.Duration
	jobCount      int
	jobCountLimit int
	jobCountLock  sync.RWMutex
	jobs          map[string]Func
	log           logger
	pollInterval  time.Duration
	queue         *Queue
}

type message struct {
	Name    string
	Message []byte
}
