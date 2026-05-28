package repo

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	urlMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url/mock"

	"github.com/stretchr/testify/assert" // 推荐使用 testify 断言，更简洁
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

var (
	ErrDBUpdateFailed = errors.New("数据库更新失败") // 统一错误实例
)

// TestDelete 表格驱动测试 Service.Delete 方法
// 命名规范：Test<方法名>，符合 Go 单测命名标准
func TestDelete(t *testing.T) {
	t.Parallel() // 无状态用例开启并发执行，提升测试效率

	// 1. 定义测试用例结构体（覆盖所有核心场景）
	type testCase struct {
		name      string                                                // 场景名称（精准描述）
		repoID    uint                                                  // 测试的仓库ID
		session   *auth.Session                                         // 会话参数
		mockSetup func(ctrl *gomock.Controller, ms *mock.MockInterface) // Mock 期望设置
		expectErr error                                                 // 预期返回的错误（nil 表示无错误）
	}

	// 2. 构建测试用例表格（整合所有场景）
	cases := []testCase{
		{
			name:   "成功-正常删除（记录存在且Delete执行成功）",
			repoID: 123,
			session: &auth.Session{
				User: &principal.User{
					ID:      1,
					Account: "actor",
				},
			},
			mockSetup: func(ctrl *gomock.Controller, ms *mock.MockInterface) {
				testCtx := context.Background()
				testRepo := &repoModel.Repo{ID: 123, Deleted: 0}

				// 期望1：GetByID 被调用且返回正常数据
				getByIDCall := ms.EXPECT().GetByID(testCtx, uint(123)).Return(testRepo, nil)
				// 期望2：Delete 在 GetByID 后被调用且返回成功
				ms.EXPECT().Delete(testCtx, uint(123)).Return(nil).After(getByIDCall)
			},
			expectErr: nil,
		},
		{
			name:    "失败-GetByID返回记录不存在（不调用Delete）",
			repoID:  456,
			session: &auth.Session{},
			mockSetup: func(ctrl *gomock.Controller, ms *mock.MockInterface) {
				testCtx := context.Background()
				expectedGetErr := gorm.ErrRecordNotFound

				// 期望1：GetByID 被调用且返回记录不存在错误
				ms.EXPECT().GetByID(testCtx, uint(456)).Return(nil, expectedGetErr)
				// 期望2：Delete 完全不被调用（增强测试严谨性）
				ms.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
			},
			expectErr: gorm.ErrRecordNotFound,
		},
		{
			name:    "失败-GetByID成功但Delete执行失败",
			repoID:  789,
			session: &auth.Session{},
			mockSetup: func(ctrl *gomock.Controller, ms *mock.MockInterface) {
				testCtx := context.Background()
				testRepo := &repoModel.Repo{ID: 789, Deleted: 0}
				expectedDeleteErr := ErrDBUpdateFailed

				// 期望1：GetByID 被调用且返回正常数据
				ms.EXPECT().GetByID(testCtx, uint(789)).Return(testRepo, nil)
				// 期望2：Delete 被调用且返回指定错误
				ms.EXPECT().Delete(testCtx, uint(789)).Return(expectedDeleteErr)
			},
			expectErr: ErrDBUpdateFailed,
		},
	}

	// 3. 遍历执行所有测试用例
	for _, tc := range cases {
		// t.Run 为每个用例创建独立的子测试，失败时可精准定位场景
		t.Run(tc.name, func(t *testing.T) {
			// 每个用例独立创建 gomock 控制器，避免用例间 Mock 污染
			ctrl := gomock.NewController(t)
			defer ctrl.Finish() // 测试结束后验证所有 Mock 期望是否执行

			// 初始化 Mock 对象
			mockRepoStore := mock.NewMockInterface(ctrl)

			// 执行当前用例的 Mock 期望设置
			tc.mockSetup(ctrl, mockRepoStore)

			// 初始化被测 Service（注入 Mock 依赖）
			service := &Service{repoStore: mockRepoStore}

			// 执行被测方法
			testCtx := context.Background()
			err := service.Delete(testCtx, tc.session, tc.repoID)

			// 4. 断言结果（使用 testify 断言更简洁，符合业界主流）
			if tc.expectErr == nil {
				assert.NoError(t, err, "Delete() 预期无错误，实际返回错误：%v", err)
			} else {
				assert.True(t, errors.Is(err, tc.expectErr),
					"Delete() 预期错误为 %v，实际为 %v", tc.expectErr, err)
			}
		})
	}
}

// 明确 RepoPageResult 类型（与业务代码对齐，泛型实例化）
type RepoPageResult = *database.PageResult[*repoModel.Repo]

// ListRepoTestCase 表格驱动用例结构体
type ListRepoTestCase struct {
	name              string
	mockRepoStoreList func(m *mock.MockInterface)   // Mock repoStore.List 行为
	mockURLProvider   func(m *urlMock.MockProvider) // Mock urlProvider 两个URL生成方法
	inputFilter       repoModel.ReqFilter           // 入参filter
	inputSession      *auth.Session                 // 入参session
	expectedRepos     []*repoModel.Repo             // 预期仓库列表
	expectedIsLast    bool                          // 预期是否最后一页
	expectedErr       error                         // 预期错误
}

// TestList 表格驱动单测入口
func TestList(t *testing.T) {
	// 1. 定义测试用例（每个用例独立配置两个Mock的行为）
	testCases := []ListRepoTestCase{
		{
			name: "场景1：正常查询 - 非最后一页",
			mockRepoStoreList: func(m *mock.MockInterface) {
				// 构造Mock分页结果
				mockPageResult := &database.PageResult[*repoModel.Repo]{
					Data: []*repoModel.Repo{
						{ID: 1, Path: "team1/project1"},
						{ID: 2, Path: "team1/project2"},
					},
					Page:      1,
					PageSize:  10,
					Total:     20,
					TotalPage: 2,
				}
				// 期望调用repoStore.List，返回Mock结果
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(mockPageResult, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 配置URL生成方法的Mock返回值，支持精确匹配Path
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team1/project1").Return("git@gitfox:team1/project1.git")
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team1/project1").Return("https://gitfox/team1/project1.git")
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team1/project2").Return("git@gitfox:team1/project2.git")
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team1/project2").Return("https://gitfox/team1/project2.git")
			},
			inputFilter:  repoModel.ReqFilter{Page: 1, PageSize: 10},
			inputSession: &auth.Session{},
			expectedRepos: []*repoModel.Repo{
				{ID: 1, Path: "team1/project1", GitSSHURL: "git@gitfox:team1/project1.git", GitURL: "https://gitfox/team1/project1.git"},
				{ID: 2, Path: "team1/project2", GitSSHURL: "git@gitfox:team1/project2.git", GitURL: "https://gitfox/team1/project2.git"},
			},
			expectedIsLast: false,
			expectedErr:    nil,
		},
		{
			name: "场景2：正常查询 - 最后一页",
			mockRepoStoreList: func(m *mock.MockInterface) {
				mockPageResult := &database.PageResult[*repoModel.Repo]{
					Data: []*repoModel.Repo{
						{ID: 3, Path: "team2/project3"},
					},
					Page:      2,
					PageSize:  10,
					Total:     15,
					TotalPage: 2,
				}
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(mockPageResult, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 配置单个仓库的URL Mock返回值
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team2/project3").Return("git@gitfox:team2/project3.git")
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team2/project3").Return("https://gitfox/team2/project3.git")
			},
			inputFilter:  repoModel.ReqFilter{Page: 2, PageSize: 10},
			inputSession: &auth.Session{},
			expectedRepos: []*repoModel.Repo{
				{ID: 3, Path: "team2/project3", GitSSHURL: "git@gitfox:team2/project3.git", GitURL: "https://gitfox/team2/project3.git"},
			},
			expectedIsLast: true,
			expectedErr:    nil,
		},
		{
			name: "场景3：repoStore.List 报错 - 错误透传",
			mockRepoStoreList: func(m *mock.MockInterface) {
				mockErr := assert.AnError
				// 期望调用List后返回空结果+错误
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, mockErr)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 存储层报错时，URLProvider方法不会被调用，无需配置期望
			},
			inputFilter:    repoModel.ReqFilter{Page: 1, PageSize: 10},
			inputSession:   &auth.Session{},
			expectedRepos:  nil,
			expectedIsLast: false,
			expectedErr:    assert.AnError,
		},
		{
			name: "场景4：空数据列表 - 无需调用URLProvider",
			mockRepoStoreList: func(m *mock.MockInterface) {
				mockPageResult := &database.PageResult[*repoModel.Repo]{
					Data:      []*repoModel.Repo{},
					Page:      1,
					PageSize:  10,
					Total:     0,
					TotalPage: 0,
				}
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(mockPageResult, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 空数据时，无需遍历填充URL，URLProvider方法不会被调用
			},
			inputFilter:    repoModel.ReqFilter{Page: 1, PageSize: 10},
			inputSession:   &auth.Session{},
			expectedRepos:  []*repoModel.Repo{},
			expectedIsLast: true,
			expectedErr:    nil,
		},
	}

	// 2. 遍历执行子测试（每个子测试独立创建Mock资源）
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ========== 每个子测试独立创建所有Mock资源 ==========
			// 1. 创建独立的Mock控制器
			ctrl := gomock.NewController(t)
			defer ctrl.Finish() // 子测试结束后自动验证Mock期望

			// 2. 创建独立的RepoStore Mock实例
			mockRepoStore := mock.NewMockInterface(ctrl)

			// 3. 创建独立的URLProvider Mock实例（核心修改：从真实实例改为Mock）
			mockURLProvider := urlMock.NewMockProvider(ctrl)

			// 4. 创建独立的Service实例（注入两个Mock依赖）
			service := &Service{
				repoStore:   mockRepoStore,
				urlProvider: mockURLProvider, // 注入URLProvider Mock
			}

			// ========== 配置当前子测试的Mock行为 ==========
			tc.mockRepoStoreList(mockRepoStore) // 配置repoStore Mock
			tc.mockURLProvider(mockURLProvider) // 配置urlProvider Mock

			// ========== 执行待测试方法 ==========
			ctx := context.Background()
			actualRepos, actualIsLast, actualErr := service.List(ctx, tc.inputSession, tc.inputFilter)

			// ========== 核心断言（精确验证所有结果） ==========
			// 1. 验证错误一致性
			assert.Equal(t, tc.expectedErr, actualErr)

			// 2. 验证是否为最后一页
			assert.Equal(t, tc.expectedIsLast, actualIsLast)

			// 3. 验证仓库列表长度一致
			assert.Equal(t, len(tc.expectedRepos), len(actualRepos))

			// 4. 验证仓库基础信息 + URL填充（精确匹配Mock返回值）
			for i := range actualRepos {
				assert.Equal(t, tc.expectedRepos[i].ID, actualRepos[i].ID)
				assert.Equal(t, tc.expectedRepos[i].Path, actualRepos[i].Path)
				assert.Equal(t, tc.expectedRepos[i].GitSSHURL, actualRepos[i].GitSSHURL)
				assert.Equal(t, tc.expectedRepos[i].GitURL, actualRepos[i].GitURL)
			}
		})
	}
}

// ====================== GetByIdentifier 方法测试 ======================
// GetByIdentifierTestCase GetByIdentifier 表格驱动用例结构体
type GetByIdentifierTestCase struct {
	name string
	// 前置条件
	mockRepoStoreGet   func(m *mock.MockInterface)   // Mock repoStore.GetBySpaceAndIdentifier
	mockURLProvider    func(m *urlMock.MockProvider) // Mock urlProvider URL生成方法
	inputSpaceID       uint                          // 入参：spaceID
	inputIdentifier    string                        // 入参：identifier
	inputAllowedStates []string                      // 入参：allowedStates
	// 预期结果
	expectedRepo *repoModel.Repo // 预期返回的repo
	expectedErr  error           // 预期错误
}

// TestService_GetByIdentifier GetByIdentifier 单测入口
func TestGetByIdentifier(t *testing.T) {
	// 1. 定义测试用例
	testCases := []GetByIdentifierTestCase{
		{
			name: "场景1：正常查询 - allowedStates为空（无需状态校验）",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				// 构造Mock repo实例
				mockRepo := &repoModel.Repo{
					ID:         1,
					SpaceID:    100,
					Identifier: "team1/project1",
					Path:       "team1/project1",
					Status:     "active",
				}
				// 期望调用GetBySpaceAndIdentifier，返回mockRepo和nil错误
				m.EXPECT().GetBySpaceAndIdentifier(
					gomock.Any(),
					uint(100),
					"team1/project1",
				).Return(mockRepo, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 配置URL生成方法返回值
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team1/project1").Return("https://gitfox/team1/project1.git")
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team1/project1").Return("git@gitfox:team1/project1.git")
			},
			inputSpaceID:       100,
			inputIdentifier:    "team1/project1",
			inputAllowedStates: []string{}, // 空状态列表
			expectedRepo: &repoModel.Repo{
				ID:         1,
				SpaceID:    100,
				Identifier: "team1/project1",
				Path:       "team1/project1",
				Status:     "active",
				GitURL:     "https://gitfox/team1/project1.git",
				GitSSHURL:  "git@gitfox:team1/project1.git",
			},
			expectedErr: nil,
		},
		{
			name: "场景2：正常查询 - allowedStates非空（repo.Status在列表中）",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:         2,
					SpaceID:    100,
					Identifier: "team1/project2",
					Path:       "team1/project2",
					Status:     "active",
				}
				m.EXPECT().GetBySpaceAndIdentifier(
					gomock.Any(),
					uint(100),
					"team1/project2",
				).Return(mockRepo, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team1/project2").Return("https://gitfox/team1/project2.git")
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team1/project2").Return("git@gitfox:team1/project2.git")
			},
			inputSpaceID:       100,
			inputIdentifier:    "team1/project2",
			inputAllowedStates: []string{"active"}, // 包含Active
			expectedRepo: &repoModel.Repo{
				ID:         2,
				SpaceID:    100,
				Identifier: "team1/project2",
				Path:       "team1/project2",
				Status:     "active",
				GitURL:     "https://gitfox/team1/project2.git",
				GitSSHURL:  "git@gitfox:team1/project2.git",
			},
			expectedErr: nil,
		},
		{
			name: "场景3：repoStore查询报错 - 错误被包装返回",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				mockErr := errors.NotFound.New("repo not found")
				// 期望调用后返回nil和错误
				m.EXPECT().GetBySpaceAndIdentifier(
					gomock.Any(),
					uint(100),
					"team1/project3",
				).Return(nil, mockErr)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 查询报错，URLProvider不会被调用，无需配置
			},
			inputSpaceID:       100,
			inputIdentifier:    "team1/project3",
			inputAllowedStates: []string{},
			expectedRepo:       nil,
			expectedErr:        errors.Wrapf(errors.NotFound.New("repo not found"), "failed to find repository: %s", "team1/project3"),
		},
		{
			name: "场景4：allowedStates非空 - repo.Status不在列表中（返回BadRequest）",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:         4,
					SpaceID:    100,
					Identifier: "team1/project4",
					Path:       "team1/project4",
					Status:     "migrate_data_import", // 状态为Disabled
				}
				m.EXPECT().GetBySpaceAndIdentifier(
					gomock.Any(),
					uint(100),
					"team1/project4",
				).Return(mockRepo, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 状态校验失败，URLProvider不会被调用，无需配置
			},
			inputSpaceID:       100,
			inputIdentifier:    "team1/project4",
			inputAllowedStates: []string{"active"}, // 仅包含Active
			expectedRepo:       nil,
			expectedErr:        errors.BadRequest.New("Repository is not ready to use."),
		},
	}

	// 2. 遍历执行子测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 每个子测试独立创建Mock资源
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 初始化Mock实例
			mockRepoStore := mock.NewMockInterface(ctrl)
			mockURLProvider := urlMock.NewMockProvider(ctrl)

			// 初始化Service
			service := &Service{
				repoStore:   mockRepoStore,
				urlProvider: mockURLProvider,
			}

			// 配置Mock行为
			tc.mockRepoStoreGet(mockRepoStore)
			tc.mockURLProvider(mockURLProvider)

			// 执行待测试方法
			ctx := context.Background()
			actualRepo, actualErr := service.GetByIdentifier(
				ctx,
				tc.inputSpaceID,
				tc.inputIdentifier,
				tc.inputAllowedStates,
			)

			// 核心断言
			// 1. 断言错误一致性
			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.Equal(t, tc.expectedErr.Error(), actualErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}

			// 2. 断言repo一致性
			assert.Equal(t, tc.expectedRepo, actualRepo)
		})
	}
}

// ====================== GetByID 方法测试 ======================
// GetByIDTestCase GetByID 表格驱动用例结构体
type GetByIDTestCase struct {
	name string
	// 前置条件
	mockRepoStoreGet func(m *mock.MockInterface)   // Mock repoStore.GetByID
	mockURLProvider  func(m *urlMock.MockProvider) // Mock urlProvider URL生成方法
	inputID          uint                          // 入参：id
	inputSession     *auth.Session                 // 入参：session
	// 预期结果
	expectedRepo *repoModel.Repo // 预期返回的repo
	expectedErr  error           // 预期错误
}

// TestService_GetByID GetByID 单测入口
func TestGetByID(t *testing.T) {
	// 1. 定义测试用例
	testCases := []GetByIDTestCase{
		{
			name: "场景1：id<1 - 返回参数非法错误",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				// id非法，不会调用repoStore，无需配置
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// id非法，不会调用URLProvider，无需配置
			},
			inputID:      0, // id<1
			inputSession: &auth.Session{},
			expectedRepo: nil,
			expectedErr:  errors.BadRequest.New("A valid repository id must be provided."),
		},
		{
			name: "场景2：正常查询 - id有效（URL回填成功）",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				mockRepo := &repoModel.Repo{
					ID:     10,
					Path:   "team1/project10",
					Status: "active",
				}
				// 期望调用GetByID，返回mockRepo和nil错误
				m.EXPECT().GetByID(gomock.Any(), uint(10)).Return(mockRepo, nil)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				m.EXPECT().GenerateGITCloneURL(gomock.Any(), "team1/project10").Return("https://gitfox/team1/project10.git")
				m.EXPECT().GenerateGITCloneSSHURL(gomock.Any(), "team1/project10").Return("git@gitfox:team1/project10.git")
			},
			inputID:      10, // 有效id
			inputSession: &auth.Session{},
			expectedRepo: &repoModel.Repo{
				ID:        10,
				Path:      "team1/project10",
				Status:    "active",
				GitURL:    "https://gitfox/team1/project10.git",
				GitSSHURL: "git@gitfox:team1/project10.git",
			},
			expectedErr: nil,
		},
		{
			name: "场景3：repoStore查询报错 - 错误被包装返回",
			mockRepoStoreGet: func(m *mock.MockInterface) {
				mockErr := errors.NotFound.New("repo not found")
				// 期望调用GetByID后返回nil和错误
				m.EXPECT().GetByID(gomock.Any(), uint(11)).Return(nil, mockErr)
			},
			mockURLProvider: func(m *urlMock.MockProvider) {
				// 查询报错，不会调用URLProvider，无需配置
			},
			inputID:      11,
			inputSession: &auth.Session{},
			expectedRepo: nil,
			expectedErr:  errors.Wrapf(errors.NotFound.New("repo not found"), "failed to find repository: %d", 11),
		},
	}

	// 2. 遍历执行子测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 每个子测试独立创建Mock资源
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 初始化Mock实例
			mockRepoStore := mock.NewMockInterface(ctrl)
			mockURLProvider := urlMock.NewMockProvider(ctrl)

			// 初始化Service
			service := &Service{
				repoStore:   mockRepoStore,
				urlProvider: mockURLProvider,
			}

			// 配置Mock行为
			tc.mockRepoStoreGet(mockRepoStore)
			tc.mockURLProvider(mockURLProvider)

			// 执行待测试方法
			ctx := context.Background()
			actualRepo, actualErr := service.GetByID(ctx, tc.inputSession, tc.inputID)

			// 核心断言
			// 1. 断言错误一致性
			if tc.expectedErr != nil {
				assert.Error(t, actualErr)
				assert.Equal(t, tc.expectedErr.Error(), actualErr.Error())
			} else {
				assert.NoError(t, actualErr)
			}

			// 2. 断言repo一致性
			assert.Equal(t, tc.expectedRepo, actualRepo)
		})
	}
}
