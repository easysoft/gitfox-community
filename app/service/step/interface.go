// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package step

import (
	"context"

	stepModel "github.com/easysoft/gitfox-community/app/model/step"
	"github.com/easysoft/gitfox-community/app/store/step"
)

var _ Interface = (*Service)(nil)

type Service struct {
	stepStore step.Interface
}

func NewStepService(stepStore step.Interface) *Service {
	return &Service{
		stepStore: stepStore,
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
