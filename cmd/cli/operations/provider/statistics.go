package provider

import (
	"github.com/google/wire"

	statisticsCtrl "github.com/easysoft/gitfox-community/app/api/controller/statistics"
	statisticsSvc "github.com/easysoft/gitfox-community/app/service/statistics"
	userStore "github.com/easysoft/gitfox-community/app/store/user"
	gitPkg "github.com/easysoft/gitfox-community/pkg/git"
)

var StatisticsWireSet = wire.NewSet(
	statisticsCtrl.NewController,
	wire.Bind(new(statisticsSvc.Interface), new(*statisticsSvc.Service)),
	statisticsSvc.NewService,
	NewStatisticsCommitCounter,
	NewStatisticsUserFinder,
)

// NewStatisticsCommitCounter 将git服务适配为统计服务需要的提交统计接口。
func NewStatisticsCommitCounter(git gitPkg.Interface) statisticsSvc.CommitCounter {
	return git
}

// NewStatisticsUserFinder 将用户存储适配为统计服务需要的用户查询接口。
func NewStatisticsUserFinder(users userStore.Interface) statisticsSvc.UserFinder {
	return users
}
