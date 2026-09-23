// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package space

import (
	"context"

	spaceModel "github.com/easysoft/gitfox-community/app/model/space"
	"github.com/easysoft/gitfox-community/pkg/database"
)

// Interface 空间数据存储接口
type Interface interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*Space], error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Space, error)
	GetByCode(ctx context.Context, code string) (*Space, error)
	GetByName(ctx context.Context, name string) (*Space, error)
	HasRepositories(ctx context.Context, spaceID uint) (bool, error)
	HasArtifacts(ctx context.Context, spaceID uint) (bool, error)
	Restore(ctx context.Context, id uint) error
}
