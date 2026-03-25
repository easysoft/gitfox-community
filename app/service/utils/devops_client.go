package utils

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/httpclient"
)

// BranchPermissionResult 分支权限返回体
type BranchPermissionResult struct {
	CreateUser    []string `json:"createUser"`
	DeleteUser    []string `json:"deleteUser"`
	UpdateUser    []string `json:"updateUser"`
	ForcePushUser []string `json:"forcePushUser"`
	SourceBranch  []string `json:"sourceBranch"`
	TargetBranch  []string `json:"targetBranch"`
	PpmCreateUser []string `json:"ppmCreateUser"`
	PpmHandleUser []string `json:"ppmHandleUser"`
	CommitLine    int      `json:"commitLine"`
	PushLine      int      `json:"pushLine"`
	ForceReview   int      `json:"forceReview"`
	ReviewFlowID  int      `json:"reviewFlowID"`
}

// GetBranchPermission 获取分支权限
func GetBranchPermission(ctx context.Context, baseURL string, repoID int64, branchName string, authorization string) (*BranchPermissionResult, error) {
	url := fmt.Sprintf("%s/branchrules?id=%d&branchName=%s", baseURL, repoID, branchName)
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := BranchPermissionResult{}
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get branch rule: %w", err)
	}
	return &result, nil
}
