package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"

	"github.com/gorhill/cronexpr"
)

const (
	// maxCronExprs is the upper bound of cron expressions per trigger.
	// Must be in sync with app/service/pipeline/trigger_cron.go
	maxCronExprs = 5
)

type Trigger struct {
	ID          int64           `json:"id"`
	RepoID      int64           `json:"repoID"`
	PipelineID  int64           `json:"pipelineID"`
	Trigger     []string        `json:"trigger"`
	Cron        []string        `json:"cron"`
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

	// 校验 cron 表达式数量
	if len(t.Cron) > maxCronExprs {
		return errors.WithI18nKey(
			errors.InvalidArgument.Newf("cron expressions count exceeds maximum limit of %d", maxCronExprs),
			i18n.PipelineTriggerParamCronLimitError,
		)
	}

	// 校验 cron 表达式，使用与调度器相同的库确保验证和执行规则一致
	if len(t.Cron) != 0 {
		for _, expr := range t.Cron {
			if _, err := cronexpr.Parse(expr); err != nil {
				return errors.WithI18nKey(errors.InvalidArgument.Newf("invalid cron expression: %s", err.Error()), i18n.PipelineTriggerParamCronError)
			}
		}
	}

	t.Trigger = slice.Unique(t.Trigger)

	return nil
}

type TriggerFilter struct {
	RepoID     int64    `json:"repoID"`
	PipelineID int64    `json:"pipelineID"`
	Trigger    []string `json:"trigger"`
}
