package codereview

import (
	"context"

	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
)

var _ Interface = (*CodeReviewService)(nil)

// Interface 代码审查服务接口
type Interface interface {
	// TriggerCodeReview 获取（或创建）内置代码审查流水线并触发一次执行
	TriggerCodeReview(ctx context.Context, in *CodeReviewInput) (*executionModel.Execution, error)
}
