package repo

import (
	"context"
	"errors"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo/mock"
	"github.com/golang/mock/gomock"
	"gorm.io/gorm"
)

// 测试用例：正常删除（记录存在，GetByID 和 Delete 均成功）
func TestServiceDeleteSuccess(t *testing.T) {
	// 1. 初始化 gomock 控制器（用于管理 Mock 生命周期）
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // 测试结束后验证所有 Mock 期望调用是否执行

	// 2. 创建 Mock 对象（基于生成的 MockRepoStoreInterface）
	mockRepoStore := mock.NewMockInterface(ctrl)

	// 3. 定义测试参数
	testID := uint(123)
	testCtx := context.Background()
	testSession := &auth.Session{
		User: &principal.User{
			ID: 1,
			Account: "actor",
		},
	} // 按需填充 Session 字段（若 service 未使用，可空）
	testRepo := &repoModel.Repo{ID: testID, Deleted: 0}

	// 4. 设置 Mock 期望行为
	// 期望 1：调用 GetByID(testCtx, testID)，返回 testRepo 和 nil 错误
	getByIDCall := mockRepoStore.EXPECT().GetByID(testCtx, testID).Return(testRepo, nil)
	// 期望 2：调用 Delete(testCtx, testID)，返回 nil 错误（必须在 GetByID 之后执行）
	mockRepoStore.EXPECT().Delete(testCtx, testID).Return(nil).After(getByIDCall)
	// 5. 初始化 service，注入 Mock 对象
	service := &Service{repoStore: mockRepoStore}

	// 6. 调用 service 层 Delete 方法
	err := service.Delete(testCtx, testSession, testID)

	// 7. 断言结果
	if err != nil {
		t.Errorf("Delete() 预期无错误，实际返回: %v", err)
	}
}

// 测试用例：GetByID 失败（记录不存在）→ service 直接返回错误，不调用 Delete
func TestServiceDeleteGetByIDFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoStore := mock.NewMockInterface(ctrl)
	testID := uint(456)
	testCtx := context.Background()
	testSession := &auth.Session{}
	expectedErr := gorm.ErrRecordNotFound // 模拟“记录不存在”错误

	// 设置 Mock 期望行为：仅调用 GetByID，返回预期错误，不调用 Delete
	mockRepoStore.EXPECT().GetByID(testCtx, testID).Return(nil, expectedErr)
	// 明确期望 Delete 不会被调用（可选，增强测试严谨性）
	mockRepoStore.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)

	// 初始化 service 并调用方法
	service := &Service{repoStore: mockRepoStore}
	err := service.Delete(testCtx, testSession, testID)

	// 断言：返回的错误与预期一致
	if !errors.Is(err, expectedErr) {
		t.Errorf("Delete() 预期错误为 %v，实际为 %v", expectedErr, err)
	}
}

// 测试用例：Delete 失败（GetByID 成功，但 Delete 执行出错）→ service 返回 Delete 错误
func TestServiceDeleteDeleteFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoStore := mock.NewMockInterface(ctrl)
	testID := uint(789)
	testCtx := context.Background()
	testSession := &auth.Session{}
	testRepo := &repoModel.Repo{ID: testID, Deleted: 0}
	expectedErr := errors.New("数据库更新失败") // 模拟 Delete 执行错误

	// 设置 Mock 期望行为
	mockRepoStore.EXPECT().GetByID(testCtx, testID).Return(testRepo, nil)
	mockRepoStore.EXPECT().Delete(testCtx, testID).Return(expectedErr)

	// 初始化 service 并调用方法
	service := &Service{repoStore: mockRepoStore}
	err := service.Delete(testCtx, testSession, testID)

	// 断言：返回的错误与 Delete 错误一致
	if !errors.Is(err, expectedErr) {
		t.Errorf("Delete() 预期错误为 %v，实际为 %v", expectedErr, err)
	}
}