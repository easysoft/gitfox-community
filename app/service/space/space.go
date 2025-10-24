package space

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	spaceDao "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"

)


type SpaceSvc struct {
	// 依赖的存储接口等，可以在这里添加
	spaceStore space.SpaceInterface
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(spaceStore space.SpaceInterface, ) *SpaceSvc {
	return &SpaceSvc{
		spaceStore: spaceStore,
	}
}

func (s *SpaceSvc) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.CreateSpaceResponse, error) {
	// Implementation logic here

	in := spaceDao.Space{
		Name:        req.Name,
		Description: req.Description,
	}
	_, err := s.spaceStore.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return &spaceModel.CreateSpaceResponse{
		ID:          "generated-id",
		Name:        req.Name,
		Description: req.Description,
	}, nil
}

