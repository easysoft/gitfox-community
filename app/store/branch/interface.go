package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
)

type Interface interface {
	// BranchType 相关接口
	BatchCreateBranchType(ctx context.Context, branchTypes []*BranchType) ([]branchModel.BranchType, error)
	UpdateBranchType(ctx context.Context, branchType *BranchType, id uint) (*branchModel.BranchType, error)
	GetBranchTypeByID(ctx context.Context, id uint) (*branchModel.BranchType, error)
	DeleteBranchType(ctx context.Context, id uint) error
	ListBranchTypes(ctx context.Context, repoID uint, filter branchModel.ReqFilterBranchType) ([]branchModel.BranchType, int64, error)
}
