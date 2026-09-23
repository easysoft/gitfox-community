// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package step

import (
	"context"

	stepModel "github.com/easysoft/gitfox-community/app/model/step"
)

// Create creates a new step
func (s *Service) Create(ctx context.Context, step *stepModel.Step) error {
	return s.stepStore.Create(ctx, step)
}

// Update updates a step
func (s *Service) Update(ctx context.Context, step *stepModel.Step) error {
	return s.stepStore.Update(ctx, step)
}

// GetByID gets step by ID
func (s *Service) GetByID(ctx context.Context, id uint) (*stepModel.Step, error) {
	return s.stepStore.GetByID(ctx, id)
}
