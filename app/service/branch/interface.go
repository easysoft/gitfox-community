package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
)

type BranchInterface interface {
	Create(ctx context.Context, req branchModel.CreateBranchRequest) (*branchModel.CreateBranchResponse, error)
}
