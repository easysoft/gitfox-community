package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ====================== Validate 方法测试 ======================
// ValidateTestCase 表格驱动用例结构体
type ValidateTestCase struct {
	name          string
	trigger       *Trigger
	expectedError bool
	expectedMsg   string
}

// TestTrigger_Validate Validate 单测入口
func TestTrigger_Validate(t *testing.T) {
	t.Parallel()

	testCases := []ValidateTestCase{
		{
			name: "场景1：有效的pipelineID和单个cron表达式",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 0 * * *"},
			},
			expectedError: false,
		},
		{
			name: "场景2：有效的pipelineID和两个cron表达式",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 0 * * *", "0 12 * * *"},
			},
			expectedError: false,
		},
		{
			name: "场景3：有效的5个cron表达式（达到限制）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron: []string{
					"0 0 * * *",
					"0 12 * * *",
					"0 0 1 * *",
					"0 0 * * 0",
					"0 15 10 * * *",
				},
			},
			expectedError: false,
		},
		{
			name: "场景4：超过限制的6个cron表达式",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron: []string{
					"0 0 * * *",
					"0 12 * * *",
					"0 0 1 * *",
					"0 0 * * 0",
					"0 15 10 * * *",
					"0 20 10 * * *", // 第 6 个，超过限制
				},
			},
			expectedError: true,
			expectedMsg:   "exceeds maximum limit",
		},
		{
			name: "场景5：有效的6字段cron（第6字段为year，?等价通配）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 10,14,16 * * * ?"},
			},
			expectedError: false,
		},
		{
			name: "场景6：有效的多个6字段cron表达式",
			trigger: &Trigger{
				PipelineID: 200,
				RepoID:     1,
				Cron:       []string{"0 15 10 * * *", "0 10,14,16 * * * ?"},
			},
			expectedError: false,
		},
		{
			name: "场景7：pipelineID无效（为0）",
			trigger: &Trigger{
				PipelineID: 0,
				Cron:       []string{"0 0 * * *"},
			},
			expectedError: true,
			expectedMsg:   "pipelineID is required",
		},
		{
			name: "场景8：pipelineID无效（负数）",
			trigger: &Trigger{
				PipelineID: -1,
				Cron:       []string{"0 0 * * *"},
			},
			expectedError: true,
			expectedMsg:   "pipelineID is required",
		},
		{
			name: "场景9：无效的cron表达式（字段数不足）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"invalid cron"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景10：无效的cron表达式（字段数过多）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 0 0 0 * * * *"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景11：无效的cron表达式（日期值无效）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 0 32 * *"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景12：无效的cron表达式（间隔从0开始）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 0/30 9-17 * *"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景13：无效的cron表达式（6字段第6字段被解析为year，MON-FRI不是合法year）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 15 10 ? * MON-FRI"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景14：空cron列表（有效）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{},
			},
			expectedError: false,
		},
		{
			name: "场景15：cron为nil（有效）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       nil,
			},
			expectedError: false,
		},
		{
			name: "场景16：有效的trigger类型和cron",
			trigger: &Trigger{
				PipelineID: 100,
				RepoID:     1,
				Trigger:    []string{"push", "pull_request"},
				Cron:       []string{"0 0 * * *"},
			},
			expectedError: false,
		},
		{
			name: "场景17：周日午夜执行（5字段）",
			trigger: &Trigger{
				PipelineID: 300,
				RepoID:     1,
				Trigger:    []string{"cron"},
				Cron:       []string{"0 0 * * 0"},
			},
			expectedError: false,
		},
		{
			name: "场景18：有效的7字段cron（含秒和逗号指定多个时间）",
			trigger: &Trigger{
				PipelineID: 400,
				RepoID:     1,
				Cron:       []string{"0 0 10,14,16 * * ? *"},
			},
			expectedError: false,
		},
		{
			name: "场景19：无效的6字段cron（误将second放首位，末尾dow值被解析为year失败）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"* 10 02 * * 2,3,6"},
			},
			expectedError: true,
			expectedMsg:   "invalid cron expression",
		},
		{
			name: "场景20：有效的7字段cron（每周二三六凌晨2:10执行）",
			trigger: &Trigger{
				PipelineID: 1,
				RepoID:     1,
				Cron:       []string{"0 10 2 * * 2,3,6 *"},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.trigger.Validate()

			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedMsg,
						"Validate() 预期错误包含 %v，实际为 %v", tc.expectedMsg, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
