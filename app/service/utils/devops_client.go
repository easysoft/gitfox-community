package utils

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	url := fmt.Sprintf("%s/branchrules?id=%d&branchName=%s", baseURL, repoID, url.QueryEscape(branchName))
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := BranchPermissionResult{}
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get branch rule: %w", err)
	}
	return &result, nil
}

// AIModelSetting AI 模型设置返回体
type AIModelSetting struct {
	Enable      bool   `json:"enable"`
	Model       string `json:"model"`
	ZaiURL      string `json:"zaiUrl"`
	ZaiToken    string `json:"zaiToken"`
	ZaiLanguage string `json:"zaiLanguage"`
}

// GetAIModelSetting 获取 AI 模型设置
func GetAIModelSetting(ctx context.Context, baseURL string, authorization string) (*AIModelSetting, error) {
	url := fmt.Sprintf("%s/ppm/aiModelSetting", baseURL)
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := AIModelSetting{}
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get ai model setting: %w", err)
	}
	return &result, nil
}

// SubmitCodeReview 提交 PPM 代码评审
func SubmitCodeReview(ctx context.Context, baseURL string, ppmID int64, createdBy string, authorization string) error {
	endpoint := fmt.Sprintf("%s/ppm/submitCodeReview?id=%d&createdBy=%s", strings.TrimRight(baseURL, "/"), ppmID, createdBy)
	headers := map[string]string{
		"Authorization": authorization,
	}
	if err := httpclient.Request(ctx, "GET", endpoint, headers, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to submit code review for PPM %d: %w", ppmID, err)
	}
	return nil
}
