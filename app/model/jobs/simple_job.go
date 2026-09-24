// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"errors"
	"time"

	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

const (
	ProgressMin = 0
	ProgressMax = 100
)

type Definition struct {
	Name       string
	Type       string
	MaxRetries int
	Timeout    time.Duration
	Data       string
}

func (def *Definition) Validate() error {
	if def.Type == "" {
		return errors.New("job Type must not be empty")
	}

	if def.Name == "" {
		return errors.New("job must have unique Name")
	}

	if def.MaxRetries < 0 {
		return errors.New("job MaxRetries must be positive")
	}

	if def.Timeout < time.Second {
		return errors.New("job Timeout too short")
	}

	return nil
}

func (def *Definition) toScheduleJob() *ScheduleJob {
	nowMilli := datetime.NowLocalTime()
	return &ScheduleJob{
		Name:                def.Name,
		Type:                def.Type,
		Priority:            PriorityNormal,
		Data:                def.Data,
		Result:              "",
		MaxDurationSeconds:  int(def.Timeout / time.Second),
		MaxRetries:          def.MaxRetries,
		State:               StateScheduled,
		Scheduled:           nowMilli,
		TotalExecutions:     0,
		RunBy:               "",
		RunDeadline:         nowMilli,
		RunProgress:         ProgressMin,
		LastExecuted:        nil,
		IsRecurring:         false,
		RecurringCron:       "",
		ConsecutiveFailures: 0,
		LastFailureError:    "",
		CreatedDate:         nowMilli,
		EditedDate:          nowMilli,
	}
}
