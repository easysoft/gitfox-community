package space

import (
	"context"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

type SpaceInterface interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error)
	Get(ctx context.Context, id uint16) (*spaceModel.SpaceResponse, error)
	Update(ctx context.Context, id uint16, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Delete(ctx context.Context, id uint16) error
}
