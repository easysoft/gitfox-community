package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
)

// Space 空间数据存储接口
type SpaceInterface interface {
	Create(ctx context.Context, req space.Space) (space.Space, error)
	Update(ctx context.Context, id uint16, req map[string]interface{}) (space.Space, error)
}
