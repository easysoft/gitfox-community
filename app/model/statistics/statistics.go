package statistics

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

const (
	// DateFormat 统计接口使用的日期格式。
	DateFormat = "2006-01-02"
	// MaxQueryDays 单次统计允许的最大天数，避免统计范围过大。
	MaxQueryDays = 366
)

// ReqForCommitStatistics 代码提交统计查询参数。
type ReqForCommitStatistics struct {
	BeginTime string `json:"beginTime" form:"beginTime" example:"2026-09-15"` // 统计开始日期，格式 yyyy-mm-dd。
	EndTime   string `json:"endTime" form:"endTime" example:"2026-09-16"`     // 统计结束日期，格式 yyyy-mm-dd。
	User      string `json:"user" form:"user" example:"admin"`                // 按人筛选的开关，传入账号或邮箱时只统计该用户，为空时统计全部用户。
	ByUser    bool   `json:"byUser" form:"byUser" example:"false"`            // 是否统计每个提交者的提交次数。
}

// Validate 校验统计查询参数。
func (req *ReqForCommitStatistics) Validate() error {
	if req.BeginTime == "" || req.EndTime == "" {
		return errors.WithI18nKey(nil, i18n.StatisticsTimeRequired)
	}

	begin, end, err := req.ParseRange()
	if err != nil {
		return err
	}
	if end.Sub(begin) > MaxQueryDays*24*time.Hour {
		return errors.WithI18nKey(nil, i18n.StatisticsRangeTooLarge)
	}

	return nil
}

// ParseRange 解析统计区间，返回左闭右闭的统计范围 [begin, end]。
func (req *ReqForCommitStatistics) ParseRange() (time.Time, time.Time, error) {
	begin, err := time.ParseInLocation(DateFormat, req.BeginTime, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errors.WithI18nKey(err, i18n.StatisticsTimeFormatInvalid)
	}

	endDate, err := time.ParseInLocation(DateFormat, req.EndTime, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errors.WithI18nKey(err, i18n.StatisticsTimeFormatInvalid)
	}
	if endDate.Before(begin) {
		return time.Time{}, time.Time{}, errors.WithI18nKey(nil, i18n.StatisticsTimeRangeInvalid)
	}

	return begin, endDate, nil
}

// DailyCount 按天统计结果。
type DailyCount struct {
	Date  string `json:"date" example:"2026-09-16"` // 统计日期。
	Count int64  `json:"count" example:"3"`         // 当日提交次数。
}

// AuthorCount 按提交者统计结果。
type AuthorCount struct {
	Account string `json:"account,omitempty" example:"admin"` // 提交者账号。
	Name    string `json:"name,omitempty" example:"张三"`       // 提交者姓名。
	Email   string `json:"email" example:"admin@example.com"` // 提交者邮箱。
	Count   int64  `json:"count" example:"3"`                 // 提交次数。
}

// RespForCommitStatistics 代码提交统计结果。
type RespForCommitStatistics struct {
	Stats   []DailyCount  `json:"stats"`             // 按日期升序排列的统计结果。
	Authors []AuthorCount `json:"authors,omitempty"` // 按提交者统计的结果，仅当 byUser 为 true 时返回。
}
