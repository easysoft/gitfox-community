// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package runner

import (
	"context"

	runnerModel "github.com/easysoft/gitfox-community/app/model/runner"
	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store runner store implementation
type Store struct {
	db         *gorm.DB
	tokenCache *TokenCache
}

// NewStore creates a new runner store instance
func NewStore(db *gorm.DB, tokenCache *TokenCache) *Store {
	return &Store{
		db:         db,
		tokenCache: tokenCache,
	}
}

type Interface interface {
	// Create creates a new runner
	Create(ctx context.Context, runner *runnerModel.Runner) error
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

	TempStoreToken(ctx context.Context, token, account string) error
	FindToken(ctx context.Context, token string) (string, error)
}
