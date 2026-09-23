// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package space

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	spaceModel "github.com/easysoft/gitfox-community/app/model/space"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/app/store/space"
	"github.com/easysoft/gitfox-community/pkg/database"
)

type Interface interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*space.Space], error)
	GetByID(ctx context.Context, id uint) (*space.Space, error)
	GetByCode(ctx context.Context, code string) (*space.Space, error)
	Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Delete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error

	GetSpaceCheckAccess(ctx context.Context, session *auth.Session, spaceCode string, permissions []constant.Permission) (*space.Space, error)
	Events(ctx context.Context, spaceID uint) (<-chan *sse.Event, <-chan error, func(context.Context) error, error)
}
