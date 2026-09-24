// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/google/wire"
)

// SystemLog 系统日志
type SystemLog struct {
	*logger.SugaredMateLogger
}

// APILog API日志
type APILog struct {
	*logger.SugaredMateLogger
}

// JobLog Job日志
type JobLog struct {
	*logger.SugaredMateLogger
}

var LogWireSet = wire.NewSet(
	NewSystemLogger,
	NewAPILogger,
	NewJobLogger,
)

func NewSystemLogger(config *config.Config) *SystemLog {
	// 初始化默认 logger，确保所有通过 logger.GetLogger(ctx) 获取的 logger
	// 都使用 lumberjack 轮转策略输出到文件（而非仅输出到 stderr）
	logger.InitDefaultLogger(config.Logging)
	// 初始化全局 logger，供 git/webhook/pipeline 等无 context 模块调用
	logger.InitGlobalLogger(config.Logging)

	return &SystemLog{
		logger.NewZapMateLogger(config.Logging, logger.LogTypeSystem).SetAsyncer(1000).Sugar(),
	}
}

func NewAPILogger(config *config.Config) *APILog {
	return &APILog{
		logger.NewZapMateLogger(config.Logging, logger.LogTypeAPIServer).SetAsyncer(1000).Sugar(),
	}
}

func NewJobLogger(config *config.Config) *JobLog {
	return &JobLog{
		logger.NewZapMateLogger(config.Logging, logger.LogTypeJob).SetAsyncer(1000).Sugar(),
	}
}
