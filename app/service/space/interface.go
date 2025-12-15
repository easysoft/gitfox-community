package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type Interface interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*space.Space], error)
	GetByID(ctx context.Context, id uint) (*space.Space, error)
	Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Delete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error

	GetSpaceCheckAccess(ctx context.Context, session *auth.Session, spaceID uint, permissions []constant.Permission) (*space.Space, error)
}
