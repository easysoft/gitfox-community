package jobs

import "time"

// 实现一个 智能的、可调整的唤醒机制 。它允许调度器在没有任务时休眠（最长 30 分钟），在有任务时按计划唤醒，而在有新任务插入时能被立即打断并重新调度，从而兼顾了实时性和资源效率。

const timerMaxDur = 30 * time.Minute
const timerMinDur = time.Nanosecond

type schedulerTimer struct {
	timerAt time.Time
	timer   *time.Timer
	// edgy 表示定时器是否为边缘触发模式。
	// 如果为 true，下一次调用 RescheduleEarlier 方法时将立即触发定时器。
	edgy bool
}

// newSchedulerTimer created new timer for the Scheduler. It is created to fire immediately.
// 创建新的定时器用于任务调度器。它被创建为立即触发。
func NewSchedulerTimer() *schedulerTimer {
	return &schedulerTimer{
		timerAt: time.Now().Add(timerMinDur),
		timer:   time.NewTimer(timerMinDur),
	}
}

// ResetAt resets the internal timer to trigger at the provided time.
// If the provided time is zero, it will schedule it to after the max duration.
// 重置内部定时器，使其在提供的时间触发。
// - 范围限制 ：它会自动调整触发时间，确保间隔既不小于 timerMinDur （防止忙轮询），也不大于 timerMaxDur （防止长时间不响应）。
// - Edgy 标志 ：设置 edgy 状态。如果为 true，表示调度器处于“急切”状态，下一次如果有新任务插入（调用 RescheduleEarlier ），定时器会立即触发，而不是等待。
func (t *schedulerTimer) ResetAt(next time.Time, edgy bool) time.Duration {
	return t.resetAt(time.Now(), next, edgy)
}

func (t *schedulerTimer) resetAt(now, next time.Time, edgy bool) time.Duration {
	var dur time.Duration

	dur = next.Sub(now)
	if dur < timerMinDur {
		dur = timerMinDur
		next = now.Add(dur)
	} else if dur > timerMaxDur {
		dur = timerMaxDur
		next = now.Add(dur)
	}

	t.Stop()
	t.edgy = edgy
	t.timerAt = next
	t.timer.Reset(dur)

	return dur
}

// RescheduleEarlier will reset the timer if the new time is earlier than the previous time.
// Otherwise, the function does nothing and returns 0.
// Providing zero time triggers the timer if it's edgy, otherwise does nothing.
// - 抢占式调度 ：当有新任务被添加到数据库时调用此方法。
// - 逻辑 ：
// - 如果新任务的执行时间 next 早于当前定时器的设定时间 t.timerAt ，则 提前 触发定时器。
// - 如果定时器处于 edgy 状态（例如之前没有待执行任务，正在空闲等待），则 立即 触发。
// - 否则，不做任何改变（保持原有的触发时间）。
func (t *schedulerTimer) RescheduleEarlier(next time.Time) time.Duration {
	return t.rescheduleEarlier(time.Now(), next)
}

func (t *schedulerTimer) rescheduleEarlier(now, next time.Time) time.Duration {
	var dur time.Duration

	switch {
	case t.edgy:
		// if the timer is edgy trigger it immediately
		dur = timerMinDur

	case next.IsZero():
		// if the provided time is zero: trigger the timer if it's edgy otherwise do nothing
		if !t.edgy {
			return 0
		}
		dur = timerMinDur

	case !next.Before(t.timerAt):
		// do nothing if the timer is already scheduled to run sooner than the provided time
		return 0

	default:
		dur = next.Sub(now)
		if dur < timerMinDur {
			dur = timerMinDur
		}
	}

	next = now.Add(dur)

	t.Stop()
	t.timerAt = next
	t.timer.Reset(dur)

	return dur
}

// 返回定时器的通道 <-chan time.Time ，供调度器的主循环 select 监听。
func (t *schedulerTimer) Ch() <-chan time.Time {
	return t.timer.C
}

// 安全地停止定时器，并清空通道中可能残留的事件（Drain channel），防止下次重置时发生误触发。
func (t *schedulerTimer) Stop() {
	// stop the timer
	t.timer.Stop()

	// consume the timer's tick if any
	select {
	case <-t.timer.C:
	default:
	}

	t.timerAt = time.Time{}
}
