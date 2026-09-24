// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package runner

import (
	"context"

	runnerModel "github.com/easysoft/gitfox-community/app/model/runner"
	"github.com/easysoft/gitfox-community/pkg/database"
)

type Interface interface {
	// Create creates a new runner
	Create(ctx context.Context, runner *runnerModel.Runner, runnerToken, runnerIP string) error
	// Update updates a runner
	Update(ctx context.Context, runner *runnerModel.Runner) error
	// GetByID gets runner by ID
	GetByID(ctx context.Context, id uint) (*runnerModel.Runner, error)
	// GetByToken gets runner by token
	GetByToken(ctx context.Context, token string) (*runnerModel.Runner, error)
	// List returns a list of runners with optional filter
	List(ctx context.Context, filter *runnerModel.ReqFilter) (*database.PageResult[*runnerModel.Runner], error)
	// Delete soft deletes a runner by ID
	Delete(ctx context.Context, id uint) error

	// GenerateRunnerToken 为 runner 生成临时访问 token
	GenerateRunnerToken(ctx context.Context, account string) (string, error)
	// ValidateRunnerToken 验证数据库中的 runner token 并返回对应的 runner
	ValidateRunnerToken(ctx context.Context, token string) (*runnerModel.Runner, error)
	// ValidateTempRunnerToken 验证缓存中的 runner token 并返回对应的 account 信息
	ValidateTempRunnerToken(ctx context.Context, token string) (string, error)
	// UpdateRunnerHeartbeat 更新 runner 心跳时间
	UpdateRunnerHeartbeat(ctx context.Context, runnerID uint) error
}
