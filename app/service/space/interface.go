package space

import (
	"context"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

type SpaceService interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.CreateSpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error)
	Get(ctx context.Context, id string) (*spaceModel.SpaceResponse, error)
	Update(ctx context.Context, id string, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Delete(ctx context.Context, id string) error
}