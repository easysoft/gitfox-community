// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pipeline

import (
	"github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/pkg/config"
)

// Controller 处理与 pipeline 相关的 HTTP 请求
type Controller struct {
	ciConfig    config.CiConfig
	pipelineSvc pipeline.Interface
}

// NewController 创建流水线处理器实例
func NewController(ciConfig config.CiConfig, pipelineSvc pipeline.Interface, repoSvc repo.Interface) *Controller {
	return &Controller{
		ciConfig:    ciConfig,
		pipelineSvc: pipelineSvc,
	}
}
