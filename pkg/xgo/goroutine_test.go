// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package xgo

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	t.Run("正常执行", func(t *testing.T) {
		var executed bool
		var wg sync.WaitGroup
		wg.Add(1)

		Go(func() {
			executed = true
			wg.Done()
		}, nil)

		wg.Wait()
		if !executed {
			t.Error("函数应该被执行")
		}
	})

	t.Run("panic恢复", func(t *testing.T) {
		var notified bool
		var wg sync.WaitGroup
		wg.Add(1)

		Go(func() {
			panic("测试panic")
		}, func(err error) {
			notified = true
			if err.Error() != "测试panic" {
				t.Errorf("期望错误消息 '测试panic', 得到 '%s'", err.Error())
			}
			wg.Done()
		})

		wg.Wait()
		if !notified {
			t.Error("应该收到panic通知")
		}
	})
}

func TestGoWithCleaner(t *testing.T) {
	t.Run("正常执行带清理函数", func(t *testing.T) {
		var executed, cleaned bool
		var wg sync.WaitGroup
		wg.Add(1)

		GoWithCleaner(func() {
			executed = true
			wg.Done()
		}, nil, func() {
			cleaned = true
		})

		wg.Wait()
		if !executed {
			t.Error("函数应该被执行")
		}
		if !cleaned {
			t.Error("清理函数应该被执行")
		}
	})

	t.Run("panic时执行清理函数", func(t *testing.T) {
		var cleaned, notified bool
		var wg sync.WaitGroup
		wg.Add(1)

		GoWithCleaner(func() {
			panic("测试panic")
		}, func(err error) {
			notified = true
			wg.Done()
		}, func() {
			cleaned = true
		})

		wg.Wait()
		if !cleaned {
			t.Error("清理函数应该被执行")
		}
		if !notified {
			t.Error("应该收到panic通知")
		}
	})

	t.Run("多个清理函数", func(t *testing.T) {
		var cleaned1, cleaned2 bool
		var wg sync.WaitGroup
		wg.Add(1)

		GoWithCleaner(func() {
			wg.Done()
		}, nil, func() {
			cleaned1 = true
		}, func() {
			cleaned2 = true
		})

		wg.Wait()
		if !cleaned1 || !cleaned2 {
			t.Error("所有清理函数都应该被执行")
		}
	})
}

func TestGoWithRetries(t *testing.T) {
	t.Run("第一次成功", func(t *testing.T) {
		var callCount int
		var wg sync.WaitGroup
		wg.Add(1)

		GoWithRetries(func() error {
			callCount++
			wg.Done()
			return nil
		}, nil, WithTimes(3))

		wg.Wait()
		if callCount != 1 {
			t.Errorf("期望调用1次, 实际调用%d次", callCount)
		}
	})

	t.Run("重试直到成功", func(t *testing.T) {
		var callCount int
		var wg sync.WaitGroup
		wg.Add(1)

		GoWithRetries(func() error {
			callCount++
			if callCount < 3 {
				return errors.New("模拟失败")
			}
			wg.Done()
			return nil
		}, nil, WithTimes(5))

		wg.Wait()
		if callCount != 3 {
			t.Errorf("期望调用3次, 实际调用%d次", callCount)
		}
	})

	t.Run("重试次数用完", func(t *testing.T) {
		var callCount int
		var wg sync.WaitGroup
		wg.Add(1)

		// 使用一个计时器来确保函数执行完成
		go func() {
			time.Sleep(200 * time.Millisecond)
			wg.Done()
		}()

		GoWithRetries(func() error {
			callCount++
			return errors.New("总是失败")
		}, nil, WithTimes(3))

		wg.Wait()
		if callCount != 3 {
			t.Errorf("期望调用3次, 实际调用%d次", callCount)
		}
	})

	t.Run("带退避策略", func(t *testing.T) {
		var callCount int
		var wg sync.WaitGroup
		wg.Add(1)
		start := time.Now()

		go func() {
			time.Sleep(300 * time.Millisecond)
			wg.Done()
		}()

		GoWithRetries(func() error {
			callCount++
			return errors.New("总是失败")
		}, nil, WithTimes(2), WithBackoff(LinearBackoff))

		wg.Wait()
		elapsed := time.Since(start)

		if callCount != 2 {
			t.Errorf("期望调用2次, 实际调用%d次", callCount)
		}
		// 应该有退避延迟，所以总时间应该大于100ms
		if elapsed < 100*time.Millisecond {
			t.Errorf("期望有退避延迟，但执行时间太短: %v", elapsed)
		}
	})
}

func TestRetryOptions(t *testing.T) {
	t.Run("WithTimes", func(t *testing.T) {
		opts := newRetryOptions()
		WithTimes(10)(opts)
		if opts.times != 10 {
			t.Errorf("期望重试次数为10, 得到%d", opts.times)
		}
	})

	t.Run("WithBackoff", func(t *testing.T) {
		opts := newRetryOptions()
		WithBackoff(RandomBackoff)(opts)
		if opts.retryMode != RandomBackoff {
			t.Errorf("期望退避模式为RandomBackoff, 得到%v", opts.retryMode)
		}
		if opts.t == nil {
			t.Error("期望ticker不为nil")
		}
	})
}

func TestBackoffMethods(t *testing.T) {
	t.Run("None模式无退避", func(t *testing.T) {
		opts := &retryOptions{retryMode: None}
		start := time.Now()
		opts.backoff()
		elapsed := time.Since(start)

		// None模式应该立即返回
		if elapsed > 10*time.Millisecond {
			t.Errorf("None模式不应该有延迟，但耗时: %v", elapsed)
		}
	})

	t.Run("LinearBackoff", func(t *testing.T) {
		opts := &retryOptions{
			retryMode: LinearBackoff,
			t:         time.NewTicker(50 * time.Millisecond),
		}
		defer opts.t.Stop()

		start := time.Now()
		opts.backoff()
		elapsed := time.Since(start)

		// 应该等待至少50ms
		if elapsed < 40*time.Millisecond {
			t.Errorf("LinearBackoff应该有延迟，但耗时太短: %v", elapsed)
		}
	})

	t.Run("RandomBackoff", func(t *testing.T) {
		opts := &retryOptions{
			retryMode: RandomBackoff,
			t:         time.NewTicker(50 * time.Millisecond),
		}
		defer opts.t.Stop()

		start := time.Now()
		opts.backoff()
		elapsed := time.Since(start)

		// RandomBackoff会重置ticker，应该有一些延迟
		if elapsed < 10*time.Millisecond {
			t.Errorf("RandomBackoff应该有延迟，但耗时太短: %v", elapsed)
		}
	})
}

func TestRecover(t *testing.T) {
	t.Run("无panic时正常执行清理", func(t *testing.T) {
		var cleaned bool

		func() {
			defer Recover(nil, func() {
				cleaned = true
			})
			// 正常执行，无panic
		}()

		if !cleaned {
			t.Error("清理函数应该被执行")
		}
	})

	t.Run("有panic时恢复并通知", func(t *testing.T) {
		var notified, cleaned bool

		func() {
			defer Recover(func(err error) {
				notified = true
				if err.Error() != "测试panic" {
					t.Errorf("期望错误消息 '测试panic', 得到 '%s'", err.Error())
				}
			}, func() {
				cleaned = true
			})
			panic("测试panic")
		}()

		if !notified {
			t.Error("应该收到panic通知")
		}
		if !cleaned {
			t.Error("清理函数应该被执行")
		}
	})

	t.Run("多个清理函数都执行", func(t *testing.T) {
		var cleaned1, cleaned2 bool

		func() {
			defer Recover(nil, func() {
				cleaned1 = true
			}, func() {
				cleaned2 = true
			})
		}()

		if !cleaned1 || !cleaned2 {
			t.Error("所有清理函数都应该被执行")
		}
	})

	t.Run("nil清理函数不会panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("nil清理函数不应该导致panic")
			}
		}()

		func() {
			defer Recover(nil)
		}()
	})
}

func TestNewRetryOptions(t *testing.T) {
	opts := newRetryOptions()

	if opts.times != defaultRetryTimes {
		t.Errorf("期望默认重试次数为%d, 得到%d", defaultRetryTimes, opts.times)
	}

	if opts.retryMode != None {
		t.Errorf("期望默认退避模式为None, 得到%v", opts.retryMode)
	}
}

// 基准测试
func BenchmarkGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		Go(func() {
			wg.Done()
		}, nil)
		wg.Wait()
	}
}

func BenchmarkGoWithRetries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		GoWithRetries(func() error {
			wg.Done()
			return nil
		}, nil, WithTimes(1))
		wg.Wait()
	}
}

// 示例测试
func ExampleGo() {
	var wg sync.WaitGroup
	wg.Add(1)

	Go(func() {
		// 不在这里输出，因为 Example 测试无法捕获 goroutine 的输出
		wg.Done()
	}, func(err error) {
		fmt.Println("Error:", err.Error())
	})

	wg.Wait()
	fmt.Println("Hello from goroutine!")
	// Output: Hello from goroutine!
}

// 修复 ExampleGo
func ExampleGoWithRetries() {
	var wg sync.WaitGroup
	wg.Add(1)
	attempt := 0

	GoWithRetries(func() error {
		attempt++
		if attempt < 3 {
			return errors.New("模拟失败")
		}
		wg.Done()
		return nil
	}, func(err error) {
		fmt.Println("最终失败:", err.Error())
	}, WithTimes(5), WithBackoff(LinearBackoff))

	wg.Wait()
	fmt.Println("成功!")
	// Output: 成功!
}
