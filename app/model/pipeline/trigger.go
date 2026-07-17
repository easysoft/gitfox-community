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
// 系统使用 gorhill/cronexpr 库解析 Cron 表达式，支持三种格式：
//
// 五字段格式:           minute hour dom month dow
// 六字段格式:           minute hour dom month dow year
// 七字段格式（推荐）:   second minute hour dom month dow year
//
// 重要：只有七字段格式才包含秒字段。六字段格式的第六个字段是 year 而非 second。
//
// 字段说明：
//   秒（second）   - 0-59，仅七字段格式的第一个字段
//   分钟（minute） - 0-59
//   小时（hour）   - 0-23
//   日期（dom）    - 1-31，支持 L（最后一天）、W（最近工作日）、?（通配）
//   月份（month）  - 1-12 或 JAN-DEC
//   星期（dow）    - 0-7 或 SUN-SAT（0和7都表示周日），支持 L、#、?（通配）
//   年份（year）   - 1970-2099，可选
//
// 支持的特殊符号：
//   * 号  - 通配符，匹配所有值
//   ? 号  - 通配符，与 * 等价（常用于 dom 和 dow 字段）
//   , 号  - 列举多个值，如 1,3,5
//   - 号  - 范围，如 MON-FRI
//   / 号  - 步长，如 */30 表示每30单位
//
// 常见示例：
//
//   1. 每天午夜执行
//      五字段: 0 0 * * *
//      七字段: 0 0 0 * * * *
//
//   2. 每天上午10点、下午2点、4点执行
//      五字段: 0 10,14,16 * * *
//      七字段: 0 0 10,14,16 * * ? *
//
//   3. 每周二、三、六凌晨2:10执行
//      五字段: 10 2 * * 2,3,6
//      七字段: 0 10 2 * * 2,3,6 *
//
//   4. 每月1号凌晨2点执行
//      五字段: 0 2 1 * *
//
//   5. 每周一上午9点执行
//      五字段: 0 9 * * 1
//
// 约束限制：
//   - 每个触发器最多支持 5 个 cron 表达式
//   - cron 表达式必须通过 gorhill/cronexpr 库的验证
//
// ========================================================================

type Trigger struct {
	ID          int64           `json:"id"`
	RepoID      int64           `json:"repoID"`
	PipelineID  int64           `json:"pipelineID"`
	Comment     string          `json:"comment"`
	Event       []string        `json:"trigger"`
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

	t.Event = slice.Unique(t.Event)

	return nil
}

type TriggerFilter struct {
	RepoID     int64    `json:"repoID"`
	PipelineID int64    `json:"pipelineID"`
	Trigger    []string `json:"trigger"`
}
