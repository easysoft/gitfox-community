package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"

	"github.com/robfig/cron/v3"
)

type Trigger struct {
	ID          int64           `json:"id"`
	PipelineID  int64           `json:"pipelineID"`
	Trigger     []string        `json:"trigger"`
	Cron        string          `json:"cron"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

// Validate 校验触发配置
func (t *Trigger) Validate() error {
	// 校验 pipelineID 是否为空
	if t.PipelineID <= 0 {
		return errors.WithI18nKey(errors.InvalidArgument.New("pipelineID is required"), i18n.PipelineTriggerParamError)
	}

	// 校验 cron 表达式
	if t.Cron != "" {
		// 使用带秒的解析器，与 CronManager 保持一致
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(t.Cron); err != nil {
			return errors.WithI18nKey(errors.InvalidArgument.Newf("invalid cron expression: %s", err.Error()), i18n.PipelineTriggerParamCronError)
		}
	}

	t.Trigger = slice.Unique(t.Trigger)

	return nil
}

type TriggerFilter struct {
	Page       int      `json:"page" form:"page" default:"1"`
	PageSize   int      `json:"pageSize" form:"pageSize" default:"20"`
	PipelineID int      `json:"pipelineID"`
	Trigger    []string `json:"trigger"`
}
