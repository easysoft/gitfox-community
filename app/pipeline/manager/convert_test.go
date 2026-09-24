// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package manager

import (
	"testing"
	"time"

	"github.com/drone/drone-go/drone"
)

func TestResolveLineTimestamp(t *testing.T) {
	base := time.Unix(1789004677, 0)
	tests := []struct {
		name string
		ts   int64
		base time.Time
		want int64
	}{
		{
			name: "相对秒数按 step 开始时间换算",
			ts:   8,
			base: base,
			want: 1789004685,
		},
		{
			name: "首行的相对时间 0 即 step 开始时间",
			ts:   0,
			base: base,
			want: 1789004677,
		},
		{
			name: "已是绝对时间的原样保留",
			ts:   1789004685,
			base: base,
			want: 1789004685,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveLineTimestamp(tt.ts, tt.base); got != tt.want {
				t.Fatalf("ResolveLineTimestamp(%d, %v) = %d, want %d", tt.ts, tt.base, got, tt.want)
			}
		})
	}
}

// 拿不到 step 开始时间时退回当前时间，避免把相对秒数当成绝对时间写进日志文件。
func TestResolveLineTimestampWithoutBase(t *testing.T) {
	start := time.Now().Unix()
	got := ResolveLineTimestamp(3, time.Time{})
	if end := time.Now().Unix(); got < start || got > end {
		t.Fatalf("ResolveLineTimestamp(3, 零值时间) = %d, 期望落在 [%d, %d] 内", got, start, end)
	}
}

func TestConvertFromDroneLine(t *testing.T) {
	base := time.Unix(1789004677, 0)
	got := ConvertFromDroneLine(&drone.Line{Number: 7, Message: "hello\n", Timestamp: 3}, base)

	if got.Number != 7 || got.Message != "hello\n" {
		t.Fatalf("日志行内容转换错误: %+v", got)
	}
	if want := int64(1789004680); got.Timestamp != want {
		t.Fatalf("Timestamp = %d, want %d", got.Timestamp, want)
	}
}
