package scan

import (
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"github.com/stretchr/testify/assert"
)

// TestNormalizeScanTriggerType 覆盖 create-by-runner 落库时的触发类型归一：
// 入参取自 GITFOX_BUILD_TRIGGER（TriggerAction 原值，或 triggerer 兜底的 "manual"）。
func TestNormalizeScanTriggerType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected constant.ScanPlanTriggerType
	}{
		{name: "空串按手动处理", input: "", expected: constant.ScanPlanTriggerTypeManual},
		{name: "manual 字面量按手动处理", input: "manual", expected: constant.ScanPlanTriggerTypeManual},
		{name: "cron 归一为 cron", input: "cron", expected: constant.ScanPlanTriggerTypeCron},
		{name: "分支事件归一为 action", input: "branch_updated", expected: constant.ScanPlanTriggerTypeAction},
		{name: "标签事件归一为 action", input: "tag_created", expected: constant.ScanPlanTriggerTypeAction},
		{name: "PR 事件归一为 action", input: "pullreq_created", expected: constant.ScanPlanTriggerTypeAction},
		{name: "未知取值兜底为 action", input: "whatever", expected: constant.ScanPlanTriggerTypeAction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, normalizeScanTriggerType(tt.input))
		})
	}
}
