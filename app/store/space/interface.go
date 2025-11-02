package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

// SpaceInterface 空间数据存储接口
type SpaceInterface interface {
	Create(ctx context.Context, req space.CreateSpaceRequest) (*space.SpaceResponse, error)
	Update(ctx context.Context, id uint, req space.UpdateSpaceRequest) (*space.SpaceResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Space, error)
	HasRepositories(ctx context.Context, spaceID uint) (bool, error)
	HasArtifacts(ctx context.Context, spaceID uint) (bool, error)
	Undelete(ctx context.Context, ids []uint, editedBy string) error
}
