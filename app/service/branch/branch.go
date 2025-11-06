package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type BranchSvc struct {
	urlProvider url.Provider
	git         git.Interface
}

// NewBranchSvc 创建 BranchSrv 实例
func NewBranchSvc(urlProvider url.Provider, git git.Interface) *BranchSvc {
	return &BranchSvc{
		urlProvider: urlProvider,
		git:         git,
	}
}

func (s *BranchSvc) Create(ctx context.Context, req branchModel.CreateBranchRequest) (*branchModel.CreateBranchResponse, error) {

	// 准备 RPC 参数
	// writeParams, err := utils.CreateRPCInternalWriteParams(c, ctl.urlProvider, session, repo)
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}

	// 调用 Git 层创建分支
	rpcOut, err := s.git.CreateBranch(ctx, &git.CreateBranchParams{
		WriteParams: writeParams,
		BranchName:  req.Name,
		Target:      req.Target,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return &branchModel.CreateBranchResponse{}, err
	}

	// 将返回结果映射为前端对象
	data, err := utils.MapBranch(rpcOut.Branch)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitMapOutputError)
		return &branchModel.CreateBranchResponse{}, err
	}
	return &data, nil
}
