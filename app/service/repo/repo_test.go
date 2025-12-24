package repo

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert" // 推荐使用 testify 断言，更简洁
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
		name        string                     // 场景名称（精准描述）
		repoID      uint                       // 测试的仓库ID
		session     *auth.Session              // 会话参数
		mockSetup   func(ctrl *gomock.Controller, ms *mock.MockInterface) // Mock 期望设置
		expectErr   error                      // 预期返回的错误（nil 表示无错误）
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
			name:   "失败-GetByID返回记录不存在（不调用Delete）",
			repoID: 456,
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
			name:   "失败-GetByID成功但Delete执行失败",
			repoID: 789,
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