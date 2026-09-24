// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package stage

import (
	"context"

	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
	"github.com/easysoft/gitfox-community/app/store/stage"
)

var _ Interface = (*Service)(nil)

type Service struct {
	stageStore stage.Interface
}

func NewStageService(stageStore stage.Interface) *Service {
	return &Service{
		stageStore: stageStore,
	}
}

type Interface interface {
	// Create creates a new stage
	Create(ctx context.Context, stage *stageModel.Stage) error
	// Update updates a stage
	Update(ctx context.Context, stage *stageModel.Stage) error
	// GetByID gets stage by ID
	GetByID(ctx context.Context, id uint) (*stageModel.Stage, error)
	GetByIDWithSteps(ctx context.Context, id uint) (*stageModel.StageWithSteps, error)
	// ListIncomplete returns a stage list where the stage is incomplete (pending or running)
	ListIncomplete(ctx context.Context) ([]*stageModel.Stage, error)
	// List returns a list of stages corresponding to an execution ID
	List(ctx context.Context, executionID uint) ([]*stageModel.Stage, error)
	// ListWithSteps returns a stage list with steps for an execution
	ListWithSteps(ctx context.Context, executionID uint) ([]*stageModel.StageWithSteps, error)
}
