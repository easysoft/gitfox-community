package pipeline

import (
	"context"
	"testing"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	pipelineStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	// TestPluginsURL 测试用插件包下载地址
	TestPluginsURL = "https://pkg.zentao.net/gitfox/20260126/plugins.zip"
)

var (
	ErrDBNotFound = errors.New("记录不存在")
	ErrDBQuery    = errors.New("数据库查询失败")
)

// ====================== GetPipelineContent 方法测试 ======================
// GetPipelineContentTestCase 表格驱动用例结构体
type GetPipelineContentTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputPipelineID   int64
	// 预期结果
	expectedResult *pipelineModel.PipelineFile
	expectedErr    error
}

// TestGetPipelineContent GetPipelineContent 单测入口
func TestGetPipelineContent(t *testing.T) {
	t.Parallel()

	testCases := []GetPipelineContentTestCase{
		{
			name: "场景1：正常查询 - 获取流水线内容成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockContent := &pipelineModel.PipelineContent{
					ID:         1,
					PipelineID: 100,
					Version:    1,
					Data:       []byte("pipeline content data"),
				}
				m.EXPECT().GetPipelineContentByPipelineID(gomock.Any(), int64(100)).Return(mockContent, nil)
			},
			inputPipelineID: 100,
			expectedResult: &pipelineModel.PipelineFile{
				Data: []byte("pipeline content data"),
			},
			expectedErr: nil,
		},
		{
			name: "场景2：流水线不存在 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetPipelineContentByPipelineID(gomock.Any(), int64(999)).Return(nil, ErrDBNotFound)
			},
			inputPipelineID: 999,
			expectedResult:  nil,
			expectedErr:     ErrDBNotFound,
		},
		{
			name: "场景3：数据库查询失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetPipelineContentByPipelineID(gomock.Any(), int64(100)).Return(nil, ErrDBQuery)
			},
			inputPipelineID: 100,
			expectedResult:  nil,
			expectedErr:     ErrDBQuery,
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
			actualResult, actualErr := service.GetPipelineContent(ctx, tc.inputPipelineID)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetPipelineContent() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== GetGroupedPlugins 方法测试 ======================
// GetGroupedPluginsTestCase 表格驱动用例结构体
type GetGroupedPluginsTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	// 预期结果
	expectedResult []*pipelineModel.GroupedPlugins
	expectedErr    error
}

// TestGetGroupedPlugins GetGroupedPlugins 单测入口
func TestGetGroupedPlugins(t *testing.T) {
	t.Parallel()

	testCases := []GetGroupedPluginsTestCase{
		{
			name: "场景1：正常查询 - 多个分组多个插件",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockPlugins := []*pipelineModel.PluginWithGroup{
					{
						Plugin: pipelineModel.Plugin{
							ID:      1,
							GroupID: 1,
							Name:    "plugin1",
							Desc:    "插件1描述",
							Logo:    "<svg>logo1</svg>",
						},
						GroupName: "构建",
						GroupDesc: "构建分组",
					},
					{
						Plugin: pipelineModel.Plugin{
							ID:      2,
							GroupID: 1,
							Name:    "plugin2",
							Desc:    "插件2描述",
							Logo:    "<svg>logo2</svg>",
						},
						GroupName: "构建",
						GroupDesc: "构建分组",
					},
					{
						Plugin: pipelineModel.Plugin{
							ID:      3,
							GroupID: 2,
							Name:    "plugin3",
							Desc:    "插件3描述",
							Logo:    "<svg>logo3</svg>",
						},
						GroupName: "部署",
						GroupDesc: "部署分组",
					},
				}
				m.EXPECT().GetAllWithGroup(gomock.Any()).Return(mockPlugins, nil)
			},
			expectedResult: []*pipelineModel.GroupedPlugins{
				{
					GroupID:   1,
					GroupName: "构建",
					Desc:      "构建分组",
					Plugins: []*pipelineModel.SimpleStep{
						{Name: "plugin1", NameCn: "插件1描述", Svg: "<svg>logo1</svg>"},
						{Name: "plugin2", NameCn: "插件2描述", Svg: "<svg>logo2</svg>"},
					},
				},
				{
					GroupID:   2,
					GroupName: "部署",
					Desc:      "部署分组",
					Plugins: []*pipelineModel.SimpleStep{
						{Name: "plugin3", NameCn: "插件3描述", Svg: "<svg>logo3</svg>"},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景2：过滤default分组 - default分组的插件不返回",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockPlugins := []*pipelineModel.PluginWithGroup{
					{
						Plugin: pipelineModel.Plugin{
							ID:      1,
							GroupID: 1,
							Name:    "plugin1",
							Desc:    "插件1描述",
							Logo:    "<svg>logo1</svg>",
						},
						GroupName: "构建",
						GroupDesc: "构建分组",
					},
					{
						Plugin: pipelineModel.Plugin{
							ID:      2,
							GroupID: 99,
							Name:    "default-plugin",
							Desc:    "默认插件",
							Logo:    "<svg>default</svg>",
						},
						GroupName: pipelineModel.DefaultPluginGroup,
						GroupDesc: "默认分组",
					},
				}
				m.EXPECT().GetAllWithGroup(gomock.Any()).Return(mockPlugins, nil)
			},
			expectedResult: []*pipelineModel.GroupedPlugins{
				{
					GroupID:   1,
					GroupName: "构建",
					Desc:      "构建分组",
					Plugins: []*pipelineModel.SimpleStep{
						{Name: "plugin1", NameCn: "插件1描述", Svg: "<svg>logo1</svg>"},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "场景3：空列表 - 无插件返回空数组",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetAllWithGroup(gomock.Any()).Return([]*pipelineModel.PluginWithGroup{}, nil)
			},
			expectedResult: []*pipelineModel.GroupedPlugins{},
			expectedErr:    nil,
		},
		{
			name: "场景4：数据库查询失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetAllWithGroup(gomock.Any()).Return(nil, ErrDBQuery)
			},
			expectedResult: nil,
			expectedErr:    ErrDBQuery,
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
			actualResult, actualErr := service.GetGroupedPlugins(ctx)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetGroupedPlugins() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== GetPluginByName 方法测试 ======================
// GetPluginByNameTestCase 表格驱动用例结构体
type GetPluginByNameTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputName         string
	// 预期结果
	expectedResult *pipelineModel.Plugin
	expectedErr    error
}

// TestGetPluginByName GetPluginByName 单测入口
func TestGetPluginByName(t *testing.T) {
	t.Parallel()

	testCases := []GetPluginByNameTestCase{
		{
			name: "场景1：正常查询 - 获取插件成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				mockPlugin := &pipelineModel.Plugin{
					ID:      1,
					GroupID: 1,
					Name:    "git-clone",
					Desc:    "Git克隆插件",
					Logo:    "<svg>logo</svg>",
					Yaml:    "kind: step\ntype: step",
					Json:    `{"name": "git-clone"}`,
					Type:    pipelineModel.PluginTypeStep,
				}
				m.EXPECT().GetPluginByName(gomock.Any(), "git-clone").Return(mockPlugin, nil)
			},
			inputName: "git-clone",
			expectedResult: &pipelineModel.Plugin{
				ID:      1,
				GroupID: 1,
				Name:    "git-clone",
				Desc:    "Git克隆插件",
				Logo:    "<svg>logo</svg>",
				Yaml:    "kind: step\ntype: step",
				Json:    `{"name": "git-clone"}`,
				Type:    pipelineModel.PluginTypeStep,
			},
			expectedErr: nil,
		},
		{
			name: "场景2：插件不存在 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetPluginByName(gomock.Any(), "non-existent").Return(nil, ErrDBNotFound)
			},
			inputName:      "non-existent",
			expectedResult: nil,
			expectedErr:    ErrDBNotFound,
		},
		{
			name: "场景3：数据库查询失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetPluginByName(gomock.Any(), "plugin").Return(nil, ErrDBQuery)
			},
			inputName:      "plugin",
			expectedResult: nil,
			expectedErr:    ErrDBQuery,
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
			actualResult, actualErr := service.GetPluginByName(ctx, tc.inputName)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.True(t, errors.Is(actualErr, tc.expectedErr),
					"GetPluginByName() 预期错误为 %v，实际为 %v", tc.expectedErr, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

// ====================== GetPluginGroupID 函数测试 ======================
// GetPluginGroupIDTestCase 表格驱动用例结构体
type GetPluginGroupIDTestCase struct {
	name string
	// 输入参数
	inputGroups    []*pipelineModel.PluginGroup
	inputGroupName string
	// 预期结果
	expectedID  int64
	expectedErr error
}

// TestGetPluginGroupID GetPluginGroupID 单测入口
func TestGetPluginGroupID(t *testing.T) {
	t.Parallel()

	testCases := []GetPluginGroupIDTestCase{
		{
			name: "场景1：匹配指定分组 - 返回对应分组ID",
			inputGroups: []*pipelineModel.PluginGroup{
				{ID: 1, Name: "构建", Desc: "构建分组"},
				{ID: 2, Name: "部署", Desc: "部署分组"},
				{ID: 3, Name: pipelineModel.DefaultPluginGroup, Desc: "默认分组"},
			},
			inputGroupName: "构建",
			expectedID:     1,
			expectedErr:    nil,
		},
		{
			name: "场景2：分组不存在 - 返回default分组ID",
			inputGroups: []*pipelineModel.PluginGroup{
				{ID: 1, Name: "构建", Desc: "构建分组"},
				{ID: 99, Name: pipelineModel.DefaultPluginGroup, Desc: "默认分组"},
			},
			inputGroupName: "不存在的分组",
			expectedID:     99,
			expectedErr:    nil,
		},
		{
			name: "场景3：分组列表为空 - 返回0",
			inputGroups:    []*pipelineModel.PluginGroup{},
			inputGroupName: "任意分组",
			expectedID:     0,
			expectedErr:    nil,
		},
		{
			name: "场景4：无default分组也无匹配分组 - 返回0",
			inputGroups: []*pipelineModel.PluginGroup{
				{ID: 1, Name: "构建", Desc: "构建分组"},
				{ID: 2, Name: "部署", Desc: "部署分组"},
			},
			inputGroupName: "测试",
			expectedID:     0,
			expectedErr:    nil,
		},
		{
			name: "场景5：匹配default分组 - 返回default分组ID",
			inputGroups: []*pipelineModel.PluginGroup{
				{ID: 1, Name: "构建", Desc: "构建分组"},
				{ID: 99, Name: pipelineModel.DefaultPluginGroup, Desc: "默认分组"},
			},
			inputGroupName: pipelineModel.DefaultPluginGroup,
			expectedID:     99,
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualID, actualErr := GetPluginGroupID(tc.inputGroups, tc.inputGroupName)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedID, actualID)
		})
	}
}

// ====================== Migrate 方法测试 ======================
// MigrateTestCase 表格驱动用例结构体
type MigrateTestCase struct {
	name string
	// 前置条件
	mockPipelineStore func(m *pipelineStoreMock.MockInterface)
	inputURL          string
	// 预期结果
	expectedErr error
}

// TestMigrate Migrate 单测入口
func TestMigrate(t *testing.T) {
	t.Parallel()

	testCases := []MigrateTestCase{
		{
			name: "场景1：正常迁移 - 新增插件成功",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				// 返回空的插件列表，表示数据库中没有插件
				m.EXPECT().GetAllPlugin(gomock.Any()).Return([]*pipelineModel.Plugin{}, nil)
				// 返回插件分组列表
				m.EXPECT().GetAllPluginGroup(gomock.Any()).Return([]*pipelineModel.PluginGroup{
					{ID: 1, Name: "源码管理", Desc: "源码管理分组"},
					{ID: 2, Name: "构建", Desc: "构建分组"},
					{ID: 3, Name: "部署", Desc: "部署分组"},
					{ID: 99, Name: pipelineModel.DefaultPluginGroup, Desc: "默认分组"},
				}, nil)
				// 允许多次创建插件调用
				m.EXPECT().CreatePlugin(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			inputURL:    TestPluginsURL,
			expectedErr: nil,
		},
		{
			name: "场景2：正常迁移 - 更新已存在的插件",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				// 返回已存在的插件列表
				m.EXPECT().GetAllPlugin(gomock.Any()).Return([]*pipelineModel.Plugin{
					{
						ID:      1,
						GroupID: 1,
						Name:    "git-clone",
						Desc:    "旧描述",
						Yaml:    "old yaml",
						Json:    "old json",
					},
				}, nil)
				// 返回插件分组列表
				m.EXPECT().GetAllPluginGroup(gomock.Any()).Return([]*pipelineModel.PluginGroup{
					{ID: 1, Name: "源码管理", Desc: "源码管理分组"},
					{ID: 99, Name: pipelineModel.DefaultPluginGroup, Desc: "默认分组"},
				}, nil)
				// 允许更新和创建插件调用
				m.EXPECT().UpdatePluginByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				m.EXPECT().CreatePlugin(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			inputURL:    TestPluginsURL,
			expectedErr: nil,
		},
		{
			name: "场景3：获取插件列表失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetAllPlugin(gomock.Any()).Return(nil, ErrDBQuery)
			},
			inputURL:    TestPluginsURL,
			expectedErr: ErrDBQuery,
		},
		{
			name: "场景4：获取插件分组失败 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetAllPlugin(gomock.Any()).Return([]*pipelineModel.Plugin{}, nil)
				m.EXPECT().GetAllPluginGroup(gomock.Any()).Return(nil, ErrDBQuery)
			},
			inputURL:    TestPluginsURL,
			expectedErr: ErrDBQuery,
		},
		{
			name: "场景5：URL为空 - 返回错误",
			mockPipelineStore: func(m *pipelineStoreMock.MockInterface) {
				m.EXPECT().GetAllPlugin(gomock.Any()).Return([]*pipelineModel.Plugin{}, nil)
			},
			inputURL:    "",
			expectedErr: errors.New("plugins url not provided"),
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
			actualErr := service.Migrate(ctx, tc.inputURL)

			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
