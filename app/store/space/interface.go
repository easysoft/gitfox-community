package space

import (
	"context"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// Interface 空间数据存储接口
type Interface interface {
	Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error)
	Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error)
	List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*Space], error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Space, error)
	GetByName(ctx context.Context, name string) (*Space, error)
	HasRepositories(ctx context.Context, spaceID uint) (bool, error)
	HasArtifacts(ctx context.Context, spaceID uint) (bool, error)
	Restore(ctx context.Context, id uint) error
}
