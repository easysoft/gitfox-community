package datetime

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/pkg/utils"
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

type TimeUnitKey string

const (
	TimeUnitSecond TimeUnitKey = "second"
	TimeUnitMinute TimeUnitKey = "minute"
	TimeUnitHour   TimeUnitKey = "hour"
	TimeUnitDay    TimeUnitKey = "day"
	TimeUnitWeek   TimeUnitKey = "week"
	TimeUnitMonth  TimeUnitKey = "month"
	TimeUnitYear   TimeUnitKey = "year"
)

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

// BetweenWithUnit returns the relative distance between two times with an appropriate unit.
func BetweenWithUnit(start, end time.Time) (int, TimeUnitKey) {
	duration := start.Sub(end)
	if duration < 0 {
		duration = -duration
	}

	seconds := ceilDivInt64(int64(duration), int64(time.Second))
	if seconds < 60 {
		return int(seconds), TimeUnitSecond
	}

	minutes := ceilDivInt64(seconds, 60)
	if minutes < 60 {
		return int(minutes), TimeUnitMinute
	}

	hours := ceilDivInt64(minutes, 60)
	if hours < 24 {
		return int(hours), TimeUnitHour
	}

	days := ceilDivInt64(hours, 24)
	if days < 7 {
		return int(days), TimeUnitDay
	}

	weeks := ceilDivInt64(days, 7)
	if weeks < 4 {
		return int(weeks), TimeUnitWeek
	}

	months := ceilDivInt64(days, 30)
	if months < 12 {
		return int(months), TimeUnitMonth
	}

	years := ceilDivInt64(days, 365)
	return int(years), TimeUnitYear
}

func ceilDivInt64(x, y int64) int64 {
	if x == 0 {
		return 0
	}
	return (x + y - 1) / y
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

// 定义东八区时区（避免硬编码偏移量，更规范）
var CST = time.FixedZone("CST", 8*3600) // UTC+8

// Local 自定义时间类型，继承 time.Time，实现 JSON 序列化/反序列化、GORM 读写接口
type Local time.Time

// MarshalJSON 实现 json.Marshaler 接口：序列化时自动转为 "2006-01-02 15:04:05" 格式
func (t *Local) MarshalJSON() ([]byte, error) {
	// 处理 nil 指针（数据库字段为 NULL 时）
	if t == nil {
		return []byte("null"), nil
	}
	// 转换为东八区时间并格式化（确保时区统一）
	formatted := time.Time(*t).In(CST).Format("2006-01-02 15:04:05")
	// 包裹双引号，符合 JSON 字符串格式
	return []byte(fmt.Sprintf("\"%s\"", formatted)), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口：反序列化（前端传参时解析）
func (t *Local) UnmarshalJSON(data []byte) error {
	// 处理 null 值
	if string(data) == "null" {
		return nil
	}
	// 解析前端传入的 "2025-11-27 14:43:19" 格式
	parsedTime, err := time.Parse(`"2006-01-02 15:04:05"`, string(data))
	if err != nil {
		return fmt.Errorf("时间格式错误，需为 2006-01-02 15:04:05: %w", err)
	}
	// 转换为东八区时间并赋值
	*t = Local(parsedTime.In(CST))
	return nil
}

// Scan 实现 gorm.Scanner 接口：数据库查询时解析时间（兼容数据库的时间类型）
func (t *Local) Scan(value interface{}) error {
	if value == nil {
		*t = Local(time.Time{})
		return nil
	}
	// 适配数据库返回的时间类型（time.Time / []byte / string）
	switch v := value.(type) {
	case time.Time:
		*t = Local(v.In(CST)) // 转换为东八区
	case []byte:
		parsedTime, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return fmt.Errorf("解析时间失败: %w", err)
		}
		*t = Local(parsedTime.In(CST))
	case string:
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("解析时间失败: %w", err)
		}
		*t = Local(parsedTime.In(CST))
	default:
		return fmt.Errorf("不支持的时间类型: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口：数据库存储时转换为 time.Time
func (t Local) Value() (driver.Value, error) {
	// 转换为标准 time.Time 类型，GORM 自动处理存储
	return time.Time(t).In(CST), nil
}

func (t *Local) String() string {
	if t == nil {
		return ""
	}
	return time.Time(*t).In(CST).Format("2006-01-02 15:04:05")
}

// Time 将项目本地时间类型转为标准 time.Time。
func (t *Local) Time() time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

// After 判断是否晚于
func (t *Local) After(t2 Local) bool {
	return time.Time(*t).After(time.Time(t2))
}

// Before 判断是否早于
func (t *Local) Before(t2 Local) bool {
	return time.Time(*t).Before(time.Time(t2))
}

// Sub 获取两个时间的时间差
func (t *Local) Sub(t2 Local) time.Duration {
	return time.Time(*t).Sub(time.Time(t2))
}

// Add 添加时间
func (t *Local) Add(d time.Duration) *Local {
	result := Local(time.Time(*t).Add(d))
	return &result
}

func NowLocalTime() *Local {
	now := time.Now().In(CST)
	localNow := Local(now)
	return &localNow
}
