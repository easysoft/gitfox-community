// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package common

import (
	"time"

	"github.com/golang-module/carbon/v2"
)

// DayStartTime 获取一天的开始时间
func DayStartTime(t carbon.Carbon) time.Time {
	return t.StartOfDay().StdTime()
}

// DayEndTime 获取一天的结束时间
func DayEndTime(t carbon.Carbon) time.Time {
	return t.EndOfDay().StdTime()
}

// DayStartUnix 获取一天的开始时间 秒
func DayStartUnix(t carbon.Carbon) int64 {
	return DayStartTime(t).Unix()
}

// DayEndUnix 获取一天的结束时间 秒
func DayEndUnix(t carbon.Carbon) int64 {
	return DayEndTime(t).Unix()
}

// DayStartUnixMilli 获取一天的开始时间 毫秒
func DayStartUnixMilli(t carbon.Carbon) int64 {
	return DayStartTime(t).UnixMilli()
}

// DayEndUnixMilli 获取一天的结束时间 毫秒
func DayEndUnixMilli(t carbon.Carbon) int64 {
	return DayEndTime(t).UnixMilli()
}

// DayStartUnixRFC3339 获取一天的开始时间 格式化 2025-01-01T00:00:00Z
func DayStartUnixRFC3339(t carbon.Carbon) string {
	return DayStartTime(t).Format(time.RFC3339)
}

// DayEndUnixRFC3339 获取一天的结束时间 格式化
func DayEndUnixRFC3339(t carbon.Carbon) string {
	return DayEndTime(t).Format(time.RFC3339)
}
