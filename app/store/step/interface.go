// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package step

import (
	"context"

	stepModel "github.com/easysoft/gitfox-community/app/model/step"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store step store implementation
type Store struct {
	db *gorm.DB
}

// NewStepStore creates a new step store instance
func NewStepStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Interface interface {
	// Create creates a new step
	Create(ctx context.Context, step *stepModel.Step) error
	// Update updates a step
	Update(ctx context.Context, step *stepModel.Step) error
	// GetByID gets step by ID
	GetByID(ctx context.Context, id uint) (*stepModel.Step, error)
}
