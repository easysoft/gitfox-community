package statistics

import (
	"context"

	statisticsModel "github.com/easysoft/gitfox-community/app/model/statistics"
)

type Interface interface {
	// GetCommitStatistics 获取按天按人统计的代码提交次数。
	GetCommitStatistics(ctx context.Context, req statisticsModel.ReqForCommitStatistics) (*statisticsModel.RespForCommitStatistics, error)
}
