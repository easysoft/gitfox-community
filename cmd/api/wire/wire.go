package wire

import (
	"context"

	branchsvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/codebase/branch"
	branchstore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/codebase/branch"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// 1. 关键修复：显式绑定接口与实现
// 告诉Wire：*branchsvc.BaseBranchService 实现了 branchsvc.BranchService 接口
var BranchInterfaceSet = wire.NewSet(
	wire.Bind(new(branchsvc.BranchService), new(*branchsvc.BaseBranchService)),
	// 提供创建实现类实例的方法（关键修复）
	branchsvc.NewBaseBranchService,
)

// 1. 定义 Provider 集合：声明依赖如何创建
// 免费版分支模块的依赖链：DB → BranchRepository → BranchService
var BranchSet = wire.NewSet(
	BranchInterfaceSet, // 引入接口绑定
	wire.Bind(new(branchstore.Repository), new(*branchstore.MySQLRepository)),
	branchstore.NewMySQLRepository,
)

// 2. 定义注入器：Wire 根据 Provider 自动生成初始化函数
// 注入免费版所有服务（返回值为需要暴露的 service 实例）
func InitPmsServices(ctx context.Context, db *gorm.DB) (
	branchsvc.BranchService,
	error,
) {
	// Wire 将自动生成代码，按依赖顺序初始化所有组件
	wire.Build(BranchSet)
	return nil, nil // 占位，实际由生成的代码替换
}
