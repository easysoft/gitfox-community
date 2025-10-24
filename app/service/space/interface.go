package space

import (
	"context"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

type SpaceService interface {
	Create(ctx context.Context, req space.CreateSpaceRequest) (*space.CreateSpaceResponse, error)
}