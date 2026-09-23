// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package xgo

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/exp/rand"
)

const (
	defaultRetryTimes = 5
	defaultInterval   = 100
)

func Go(fn func(), notifyFunc func(error)) {
	go goSafe(fn, notifyFunc)
}

func GoWithCleaner(fn func(), notifyFunc func(error), cleaner ...func()) {
	go goSafe(fn, notifyFunc, cleaner...)
}

func goSafe(fn func(), notifyFunc func(error), cleaner ...func()) {
	defer Recover(notifyFunc, cleaner...)
	fn()
}

type (
	RetryOption  func(*retryOptions)
	retryOptions struct {
		retryMode RetryMode
		times     int
		t         *time.Ticker
	}
)

// 退避策略：线性退避、随机退避、指数退避
type RetryMode int

const (
	None RetryMode = iota
	LinearBackoff
	RandomBackoff
	ExponentialBackoff
)

func WithTimes(times int) RetryOption {
	return func(options *retryOptions) {
		options.times = times
	}
}
func WithBackoff(mode RetryMode) RetryOption {
	return func(options *retryOptions) {
		options.retryMode = mode
		options.t = time.NewTicker(defaultInterval * time.Millisecond)
	}
}

func GoWithRetries(fn func() error, notifyFunc func(error), opts ...RetryOption) {
	var options = newRetryOptions()
	for _, opt := range opts {
		opt(options)
	}
	go goSafe(func() {
		for i := 0; i < options.times; i++ {
			var err error
			if err = fn(); err == nil {
				return
			}
			fmt.Fprintf(os.Stderr, "GoWithRetries err %+v \n", err)
			options.backoff()
		}
	}, notifyFunc)

}

func newRetryOptions() *retryOptions {
	rand.Seed(uint64(time.Now().UTC().UnixNano()))
	return &retryOptions{
		times:     defaultRetryTimes,
		retryMode: None,
	}
}

func (options *retryOptions) backoff() {
	if options.retryMode == None {
		return
	}
	<-options.t.C
	switch options.retryMode {
	case LinearBackoff:
	case RandomBackoff:
		//随机粒度还是要调用端按需传参的（ms/s/min），retryOptions中增加基数字段
		options.t.Reset(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func Recover(notifyFunc func(error), cleanups ...func()) {
	if cleanups != nil {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}

	if err := recover(); err != nil {
		if notifyFunc != nil {
			notifyFunc(errors.New(fmt.Sprint(err)))
		}
		fmt.Fprintf(os.Stderr, "time:%d,  recover occurs %+v , stack:%s \n", time.Now().UTC().UnixNano(), err, string(debug.Stack()))
	}
}
