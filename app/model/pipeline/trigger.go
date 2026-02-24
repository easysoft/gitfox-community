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

// ========================================================================
// Cron 表达式格式说明文档
// ========================================================================
//
// 系统支持 Quartz Cron 表达式格式（使用 gorhill/cronexpr 库解析）
// 支持标准的五字段或六字段格式
//
// 六字段格式（推荐）: second minute hour day month weekday
// 五字段格式:         minute hour day month weekday
//
// 字段说明（从左到右）：
//   第一字段（秒）        - 0-59，可选（仅六字段格式）
//   第一/二字段（分钟）   - 0-59
//   第二/三字段（小时）   - 0-23
//   第三/四字段（日期）   - 1-31 或 ? （? 表示不指定具体日期，与 weekday 互斥）
//   第四/五字段（月份）   - 1-12 或 JAN-DEC
//   第五/六字段（星期）   - 0-7 或 SUN-SAT （0和7都表示周日），或 ? （与 day 互斥）
//
// 支持的特殊符号：
//   * 号        - 表示任何时刻/日期/月份/星期（通配符）
//   ? 号        - 表示不指定值，仅用于 day 和 weekday 字段（两者必须有一个为 ?）
//   , 号        - 表示列举多个值，如 10,14,16 表示 10点、14点、16点
//   - 号        - 表示范围，如 MON-FRI 表示周一到周五
//   / 号        - 表示间隔/步长，如 0/30 表示从0开始每30分钟，9-17/1 表示9-17点每1小时
//
// 常见示例（都是有效格式）：
//
//   1. 每天午夜执行（两种格式都支持）
//      标准五字段: 0 0 * * *
//      六字段格式: 0 0 0 * * ?
//
//   2. 每天上午10点、下午2点、4点执行
//      格式: 0 0 10,14,16 * * ?
//
//   3. 工作日（周一至周五）上午10:15执行
//      格式: 0 15 10 ? * MON-FRI
//
//   4. 9点到17点每30分钟执行一次
//      格式: 0 0/30 9-17 * * ?
//
//   5. 每月1号凌晨2点执行
//      格式: 0 2 1 * * ?
//
//   6. 每周一上午9点执行
//      格式: 0 0 9 ? * MON
//
//   7. 每年1月1日执行
//      格式: 0 0 0 1 1 ?
//
//   8. 每工作日（周一-周五）上午10:30执行
//      格式: 0 30 10 ? * 1-5
//
// 约束限制：
//   - 每个触发最多支持 5 个 cron 表达式
//   - day 和 weekday 字段必须有一个为 ? （不能同时为 *）
//   - cron 表达式必须通过 gorhill/cronexpr 库的验证
//
// ========================================================================

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
