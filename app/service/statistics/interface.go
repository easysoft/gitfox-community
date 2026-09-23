// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      liyang <liyang@easycorp.ltd>

package statistics

import (
	"context"

	statisticsModel "github.com/easysoft/gitfox-community/app/model/statistics"
)

type Interface interface {
	// GetCommitStatistics 获取按天按人统计的代码提交次数。
	GetCommitStatistics(ctx context.Context, req statisticsModel.ReqForCommitStatistics) (*statisticsModel.RespForCommitStatistics, error)
}
