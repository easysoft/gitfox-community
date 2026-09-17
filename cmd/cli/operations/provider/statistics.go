package provider

import (
	"github.com/google/wire"

	statisticsCtrl "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/statistics"
	statisticsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/statistics"
	statisticsStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/statistics"
	userStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	gitPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
)

var StatisticsWireSet = wire.NewSet(
	statisticsCtrl.NewController,
	wire.Bind(new(statisticsSvc.Interface), new(*statisticsSvc.Service)),
	statisticsSvc.NewService,
	NewStatisticsCommitCounter,
	NewStatisticsUserFinder,
	wire.Bind(new(statisticsStore.Interface), new(*statisticsStore.Store)),
	statisticsStore.NewStore,
)

// NewStatisticsCommitCounter 将git服务适配为统计服务需要的提交统计接口。
func NewStatisticsCommitCounter(git gitPkg.Interface) statisticsSvc.CommitCounter {
	return git
}

// NewStatisticsUserFinder 将用户存储适配为统计服务需要的用户查询接口。
func NewStatisticsUserFinder(users userStore.Interface) statisticsSvc.UserFinder {
	return users
}
