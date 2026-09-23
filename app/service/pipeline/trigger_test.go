// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"context"
	"testing"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	pipelineStoreMock "github.com/easysoft/gitfox-community/app/store/pipeline/mock"
	"github.com/easysoft/gitfox-community/pkg/errors"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ====================== CreateTrigger 方法测试 ======================
// CreateTriggerTestCase 表格驱动用例结构体
type CreateTriggerTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputTrigger      *pipelineModel.Trigger
	// 预期结果
	expectedErr error
}

// TestCreateTrigger CreateTrigger 单测入口
func TestCreateTrigger(t *testing.T) {
	t.Parallel()

	testCases := []CreateTriggerTestCase{
		{
			name: "场景1：正常创建 - 创建触发成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Event:      []string{"push", "pull_request"},
				Cron:       []string{"0 0 * * *"},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景2：数据库创建失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(ErrDBQuery)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Event:      []string{"push"},
				CreatedBy:  "admin",
			},
			expectedErr: ErrDBQuery,
		},
		{
			name: "场景3：含逗号的cron表达式 - 创建成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Event:      []string{"push"},
				Cron:       []string{"0 0 10,14,16 * * ?"},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景4：多条cron表达式 - 创建成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 200,
				Event:      []string{"cron"},
				Cron: []string{
					"0 15 10 ? * MON-FRI",
					"0 0 10,14,16 * * ?",
					"0 0/30 9-17 * * ?",
				},
				CreatedBy: "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景5：仅cron无trigger - 创建成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Cron:       []string{"0 0 * * *"},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景6：仅trigger无cron - 创建成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Event:      []string{"push", "tag"},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景7：空cron数组 - 创建成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				PipelineID: 100,
				Event:      []string{"push"},
				Cron:       []string{},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineStore := pipelineStoreMock.NewMockInterface(ctrl)

			service := &Service{
				pipelineStore: mockPipelineStore,
			}

			tc.mockPipelineStore(mockPipelineStore)

			ctx := context.Background()
			actualErr := service.CreateTrigger(ctx, tc.inputTrigger)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"CreateTrigger() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// ====================== UpdateTrigger 方法测试 ======================
// UpdateTriggerTestCase 表格驱动用例结构体
type UpdateTriggerTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputTrigger      *pipelineModel.Trigger
	// 预期结果
	expectedErr error
}

// TestUpdateTrigger UpdateTrigger 单测入口
func TestUpdateTrigger(t *testing.T) {
	t.Parallel()

	testCases := []UpdateTriggerTestCase{
		{
			name: "场景1：正常更新 - 更新触发成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{"push", "tag"},
				Cron:       []string{"0 12 * * *"},
				EditedBy:   "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景2：数据库更新失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Any()).Return(ErrDBQuery)
			},
			inputTrigger: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{"push"},
				EditedBy:   "admin",
			},
			expectedErr: ErrDBQuery,
		},
		{
			name: "场景3：更新cron为含逗号的表达式 - 更新成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{"push"},
				Cron:       []string{"0 0 10,14,16 * * ?", "0 0/30 9-17 * * ?"},
				EditedBy:   "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景4：清空cron保留trigger - 更新成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{"push"},
				Cron:       []string{},
				EditedBy:   "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景5：cron和trigger都为空 - store层会硬删除",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Any()).Return(nil)
			},
			inputTrigger: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{},
				Cron:       []string{},
				EditedBy:   "admin",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineStore := pipelineStoreMock.NewMockInterface(ctrl)

			service := &Service{
				pipelineStore: mockPipelineStore,
			}

			tc.mockPipelineStore(mockPipelineStore)

			ctx := context.Background()
			actualErr := service.UpdateTrigger(ctx, tc.inputTrigger)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"UpdateTrigger() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// ====================== GetTriggerByID 方法测试 ======================
// GetTriggerByIDTestCase 表格驱动用例结构体
type GetTriggerByIDTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputID           int64
	// 预期结果
	expectedResult *pipelineModel.Trigger
	expectedErr    error
}

// TestGetTriggerByID GetTriggerByID 单测入口
func TestGetTriggerByID(t *testing.T) {
	t.Parallel()

	testCases := []GetTriggerByIDTestCase{
		{
			name: "场景1：正常查询 - 获取触发成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockTrigger := &pipelineModel.Trigger{
					ID:         1,
					PipelineID: 100,
					Event:      []string{"push", "pull_request"},
					Cron:       []string{"0 0 * * *"},
					CreatedBy:  "admin",
				}
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(1)).Return(mockTrigger, nil)
			},
			inputID: 1,
			expectedResult: &pipelineModel.Trigger{
				ID:         1,
				PipelineID: 100,
				Event:      []string{"push", "pull_request"},
				Cron:       []string{"0 0 * * *"},
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景2：触发不存在 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(999)).Return(nil, ErrDBNotFound)
			},
			inputID:        999,
			expectedResult: nil,
			expectedErr:    ErrDBNotFound,
		},
		{
			name: "场景3：数据库查询失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(1)).Return(nil, ErrDBQuery)
			},
			inputID:        1,
			expectedResult: nil,
			expectedErr:    ErrDBQuery,
		},
		{
			name: "场景4：查询含逗号cron表达式的触发 - 返回正确数据",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockTrigger := &pipelineModel.Trigger{
					ID:         2,
					PipelineID: 200,
					Event:      []string{"cron"},
					Cron: []string{
						"0 0 10,14,16 * * ?",
						"0 0/30 9-17 * * ?",
						"0 15 10 ? 6L 2002-2006",
					},
					CreatedBy: "admin",
				}
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(2)).Return(mockTrigger, nil)
			},
			inputID: 2,
			expectedResult: &pipelineModel.Trigger{
				ID:         2,
				PipelineID: 200,
				Event:      []string{"cron"},
				Cron: []string{
					"0 0 10,14,16 * * ?",
					"0 0/30 9-17 * * ?",
					"0 15 10 ? 6L 2002-2006",
				},
				CreatedBy: "admin",
			},
			expectedErr: nil,
		},
		{
			name: "场景5：查询cron为空的触发 - 返回空cron数组",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockTrigger := &pipelineModel.Trigger{
					ID:         3,
					PipelineID: 300,
					Event:    []string{"push"},
					Cron:       nil,
					CreatedBy:  "admin",
				}
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(3)).Return(mockTrigger, nil)
			},
			inputID: 3,
			expectedResult: &pipelineModel.Trigger{
				ID:         3,
				PipelineID: 300,
				Event:    []string{"push"},
				Cron:       nil,
				CreatedBy:  "admin",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineStore := pipelineStoreMock.NewMockInterface(ctrl)

			service := &Service{
				pipelineStore: mockPipelineStore,
			}

			tc.mockPipelineStore(mockPipelineStore)

			ctx := context.Background()
			actualResult, actualErr := service.GetTriggerByID(ctx, tc.inputID)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetTriggerByID() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== DeleteTrigger 方法测试 ======================
// DeleteTriggerTestCase 表格驱动用例结构体
type DeleteTriggerTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputID           int64
	// 预期结果
	expectedErr error
}

// TestDeleteTrigger DeleteTrigger 单测入口
func TestDeleteTrigger(t *testing.T) {
	t.Parallel()

	testCases := []DeleteTriggerTestCase{
		{
			name: "场景1：正常删除 - 删除触发成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(1)).Return(&pipelineModel.Trigger{
					ID:         1,
					PipelineID: 100,
					Event:    []string{"push"},
					Cron:       []string{},
				}, nil)
				m.EXPECT().DeleteTrigger(gomock.Any(), int64(1)).Return(nil)
			},
			inputID:     1,
			expectedErr: nil,
		},
		{
			name: "场景2：触发不存在 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(999)).Return(nil, ErrDBNotFound)
			},
			inputID:     999,
			expectedErr: ErrDBNotFound,
		},
		{
			name: "场景3：数据库删除失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggerByID(gomock.Any(), int64(1)).Return(&pipelineModel.Trigger{
					ID:         1,
					PipelineID: 100,
				}, nil)
				m.EXPECT().DeleteTrigger(gomock.Any(), int64(1)).Return(ErrDBQuery)
			},
			inputID:     1,
			expectedErr: ErrDBQuery,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineStore := pipelineStoreMock.NewMockInterface(ctrl)

			service := &Service{
				pipelineStore: mockPipelineStore,
			}

			tc.mockPipelineStore(mockPipelineStore)

			ctx := context.Background()
			actualErr := service.DeleteTrigger(ctx, tc.inputID)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"DeleteTrigger() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// ====================== GetTriggers 方法测试 ======================
// GetTriggersTestCase 表格驱动用例结构体
type GetTriggersTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputFilter       *pipelineModel.TriggerFilter
	// 预期结果
	expectedResult []*pipelineModel.Trigger
	expectedErr    error
}

// TestGetTriggers GetTriggers 单测入口
func TestGetTriggers(t *testing.T) {
	t.Parallel()

	testCases := []GetTriggersTestCase{
		{
			name: "场景1：无过滤条件 - 获取全部触发列表",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{
					{
						ID:         1,
						PipelineID: 100,
						Event:    []string{"push"},
						Cron:       []string{"0 0 * * *"},
						CreatedBy:  "admin",
					},
					{
						ID:         2,
						PipelineID: 101,
						Event:    []string{"tag"},
						Cron:       []string{"0 0 10,14,16 * * ?"},
						CreatedBy:  "user1",
					},
				}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter: &pipelineModel.TriggerFilter{},
			expectedResult: []*pipelineModel.Trigger{
				{
					ID:         1,
					PipelineID: 100,
					Event:    []string{"push"},
					Cron:       []string{"0 0 * * *"},
					CreatedBy:  "admin",
				},
				{
					ID:         2,
					PipelineID: 101,
					Event:    []string{"tag"},
					Cron:       []string{"0 0 10,14,16 * * ?"},
					CreatedBy:  "user1",
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景2：按流水线ID过滤 - 获取触发列表成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{
					{
						ID:         1,
						PipelineID: 100,
						Event:    []string{"push"},
						Cron:       []string{"0 0 * * *"},
						CreatedBy:  "admin",
					},
				}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter: &pipelineModel.TriggerFilter{
				PipelineID: 100,
			},
			expectedResult: []*pipelineModel.Trigger{
				{
					ID:         1,
					PipelineID: 100,
					Event:    []string{"push"},
					Cron:       []string{"0 0 * * *"},
					CreatedBy:  "admin",
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景3：按仓库ID过滤 - 获取触发列表成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{
					{
						ID:         1,
						RepoID:     10,
						PipelineID: 100,
						Event:    []string{"push"},
						Cron:       []string{"0 0/30 9-17 * * ?"},
						CreatedBy:  "admin",
					},
				}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter: &pipelineModel.TriggerFilter{
				RepoID: 10,
			},
			expectedResult: []*pipelineModel.Trigger{
				{
					ID:         1,
					RepoID:     10,
					PipelineID: 100,
					Event:    []string{"push"},
					Cron:       []string{"0 0/30 9-17 * * ?"},
					CreatedBy:  "admin",
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景4：按触发类型过滤 - 获取触发列表成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{
					{
						ID:         1,
						PipelineID: 100,
						Event:    []string{"push"},
						Cron:       []string{"0 0 * * *"},
						CreatedBy:  "admin",
					},
				}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter: &pipelineModel.TriggerFilter{
				Trigger: []string{"push"},
			},
			expectedResult: []*pipelineModel.Trigger{
				{
					ID:         1,
					PipelineID: 100,
					Event:    []string{"push"},
					Cron:       []string{"0 0 * * *"},
					CreatedBy:  "admin",
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景5：空列表 - 返回空数组",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter:    &pipelineModel.TriggerFilter{},
			expectedResult: []*pipelineModel.Trigger{},
			expectedErr:    nil,
		},
		{
			name: "场景6：数据库查询失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(nil, ErrDBQuery)
			},
			inputFilter:    &pipelineModel.TriggerFilter{},
			expectedResult: nil,
			expectedErr:    ErrDBQuery,
		},
		{
			name: "场景7：返回含多条复杂cron表达式的触发",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockResult := []*pipelineModel.Trigger{
					{
						ID:         1,
						PipelineID: 100,
						Event:    []string{"cron"},
						Cron: []string{
							"0 15 10 ? * MON-FRI",
							"0 0 10,14,16 * * ?",
							"0 0/30 9-17 * * ?",
						},
						CreatedBy: "admin",
					},
				}
				m.EXPECT().GetTriggers(gomock.Any(), gomock.Any()).Return(mockResult, nil)
			},
			inputFilter: &pipelineModel.TriggerFilter{PipelineID: 100},
			expectedResult: []*pipelineModel.Trigger{
				{
					ID:         1,
					PipelineID: 100,
					Event:    []string{"cron"},
					Cron: []string{
						"0 15 10 ? * MON-FRI",
						"0 0 10,14,16 * * ?",
						"0 0/30 9-17 * * ?",
					},
					CreatedBy: "admin",
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineStore := pipelineStoreMock.NewMockInterface(ctrl)

			service := &Service{
				pipelineStore: mockPipelineStore,
			}

			tc.mockPipelineStore(mockPipelineStore)

			ctx := context.Background()
			actualResult, actualErr := service.GetTriggers(ctx, tc.inputFilter)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetTriggers() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
