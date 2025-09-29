package router

import "gitfox.zcorp.cc/git/demo/gitfox-pms/app/service/codebase/branch"

// ServiceContainer 服务容器，存储所有服务实例
type ServiceContainer struct {
	BranchService   branch.BranchService   // 分支服务
	// 后续新增服务只需在此添加字段，如：
	// ScanService scan.Service
}

// NewServiceContainer 创建服务容器实例
func NewServiceContainer(
	branchSvc branch.BranchService,
	// 后续新增服务只需添加参数，如：scanSvc scan.Service
) *ServiceContainer {
	return &ServiceContainer{
		BranchService:   branchSvc,
		// 后续新增服务：ScanService: scanSvc
	}
}