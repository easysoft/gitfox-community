package space

import (
	"context"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
)

type SpaceSvc struct {
	// 依赖的存储接口等，可以在这里添加
	spaceStore space.SpaceInterface
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(spaceStore space.SpaceInterface) *SpaceSvc {
	return &SpaceSvc{
		spaceStore: spaceStore,
	}
}

func (s *SpaceSvc) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 CreateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Create(ctx, req)
}

func (s *SpaceSvc) List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Get(ctx context.Context, id uint) (*spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// 直接将 UpdateSpaceRequest 传递给 store 层，由 store 层进行数据转换并返回 SpaceResponse
	return s.spaceStore.Update(ctx, id, req)
}

func (s *SpaceSvc) Delete(ctx context.Context, id uint) error {
	// Implementation logic here
	return nil
}
