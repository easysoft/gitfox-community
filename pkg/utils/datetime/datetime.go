package datetime

import (
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils"
)

var timeFormat map[string]string

func init() {
	timeFormat = map[string]string{
		// rfc3339Milli就像时间rfc3339纳米级，但具有毫秒级精度，并且小数秒没有拖尾
		"RFC3339Milli":        "2006-01-02T15:04:05.000Z07:00",
		"yyyy-mm-dd hh:mm:ss": "2006-01-02 15:04:05",
		"yyyy-mm-dd hh:mm":    "2006-01-02 15:04",
		"yyyy-mm-dd hh":       "2006-01-02 15",
		"yyyy-mm-dd":          "2006-01-02",
		"yyyy-mm":             "2006-01",
		"mm-dd":               "01-02",
		"dd-mm-yy hh:mm:ss":   "02-01-06 15:04:05",
		"yyyy/mm/dd hh:mm:ss": "2006/01/02 15:04:05",
		"yyyy/mm/dd hh:mm":    "2006/01/02 15:04",
		"yyyy/mm/dd hh":       "2006/01/02 15",
		"yyyy/mm/dd":          "2006/01/02",
		"yyyy/mm":             "2006/01",
		"mm/dd":               "01/02",
		"dd/mm/yy hh:mm:ss":   "02/01/06 15:04:05",
		"yyyymmdd":            "20060102",
		"mmddyy":              "010206",
		"yyyy":                "2006",
		"yy":                  "06",
		"mm":                  "01",
		"hh:mm:ss":            "15:04:05",
		"hh:mm":               "15:04",
		"mm:ss":               "04:05",
	}
}

// AddDuration add or sub duration to the time.
func AddDuration(t time.Time, duration time.Duration) time.Time {
	return t.Add(duration)
}

// AddMinute add or sub minutes to the time.
func AddMinute(t time.Time, minutes int64) time.Time {
	return t.Add(time.Minute * time.Duration(minutes))
}

// AddHour add or sub hours to the time.
func AddHour(t time.Time, hours int64) time.Time {
	return t.Add(time.Hour * time.Duration(hours))
}

// AddDay add or sub days to the time.
func AddDay(t time.Time, days int64) time.Time {
	return t.Add(24 * time.Hour * time.Duration(days))
}

// AddWeek add or sub weeks to the time.
func AddWeek(t time.Time, weeks int64) time.Time {
	return t.Add(7 * 24 * time.Hour * time.Duration(weeks))
}

// AddMonth add or sub months to the time.
func AddMonth(t time.Time, months int64) time.Time {
	return t.AddDate(0, int(months), 0)
}

// AddYear add or sub year to the time.
func AddYear(t time.Time, year int64) time.Time {
	return t.AddDate(int(year), 0, 0)
}

// GetNowDate 返回当前日期的格式yyyy-mm-dd。
func GetNowDate() string {
	return time.Now().Format("2006-01-02")
}

// GetNowTime 返回当前时间的格式hh-mm-ss。
func GetNowTime() string {
	return time.Now().Format("15:04:05")
}

// GetNowDateTime 返回当前日期时间的格式yyyy-mm-dd hh-mm-ss。
func GetNowDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// FormatTimeToStr 将时间转换为字符串。
func FormatTimeToStr(t time.Time, format string, timezone ...string) string {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return ""
	}

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return ""
		}
		return t.In(loc).Format(tf)
	}
	return t.Format(tf)
}

// FormatStrToTime 将字符串转换为时间。
func FormatStrToTime(str, format string, timezone ...string) (time.Time, error) {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return time.Time{}, fmt.Errorf("format %s not support", format)
	}

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return time.Time{}, err
		}

		return time.ParseInLocation(tf, str, loc)
	}

	return time.Parse(tf, str)
}

// NowDateOrTime 返回具有特定格式和时区的当前日期时间。
func NowDateOrTime(format string, timezone ...string) string {
	tf, ok := timeFormat[strings.ToLower(format)]
	if !ok {
		return ""
	}

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return ""
		}

		return time.Now().In(loc).Format(tf)
	}

	return time.Now().Format(tf)
}

// Timestamp 返回当前秒级时间戳。
func Timestamp(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}

		t = t.In(loc)
	}

	return t.Unix()
}

// TimestampMilli 返回当前毫秒级时间戳。
func TimestampMilli(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
}

// TimestampMicro 返回当前微秒级时间戳。
func TimestampMicro(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return int64(time.Nanosecond) * t.UnixNano() / int64(time.Microsecond)
}

// TimestampNano 返回当前纳秒级时间戳。
func TimestampNano(timezone ...string) int64 {
	t := time.Now()

	if timezone != nil && timezone[0] != "" {
		loc, err := time.LoadLocation(timezone[0])
		if err != nil {
			return 0
		}
		t = t.In(loc)
	}

	return t.UnixNano()
}

// TrackFuncTime 跟踪函数执行时间。
// 调用它在函数顶部，如 `defer TrackFuncTime(time.Now())()`
func TrackFuncTime(pre time.Time) func() {
	callerName := utils.GetCallerName()
	return func() {
		elapsed := time.Since(pre)
		fmt.Printf("Function %s execution time:\t %v", callerName, elapsed)
	}
}
