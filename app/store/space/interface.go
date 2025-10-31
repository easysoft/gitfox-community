package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

// SpaceInterface 空间数据存储接口
type SpaceInterface interface {
	Create(ctx context.Context, req space.CreateSpaceRequest) (*space.SpaceResponse, error)
	Update(ctx context.Context, id uint, req space.UpdateSpaceRequest) (*space.SpaceResponse, error)
}
