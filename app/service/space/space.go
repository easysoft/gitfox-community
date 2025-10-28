package space

import (
	"context"

	spaceDao "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
	// 从 context 中获取 logger
	log := logger.GetLogger(ctx)
	log.Debugf("[Service] 开始处理创建空间业务逻辑, space_name=%s", req.Name)

	in := spaceDao.Space{
		Name:      req.Name,
		Desc:      req.Desc,
		CreatedBy: "testUser",
	}

	log.Debugf("[Service] 准备调用 Store 层创建空间")

	out, err := s.spaceStore.Create(ctx, in)
	if err != nil {
		log.Errorf("[Service] Store 层创建空间失败: %v", err)
		return nil, err
	}

	log.Debugf("[Service] 空间创建成功, space_name=%s", out.Name)

	return &spaceModel.SpaceResponse{
		ID:          out.ID,
		Name:        req.Name,
		Owner:       req.Owner,
		Permission:  req.Permission,
		Desc:        req.Desc,
		CreatedBy:   req.CreatedBy,
		CreatedDate: out.CreatedDate,
		EditedBy:    out.EditedBy,
		EditedDate:  out.EditedDate,
	}, nil
}

func (s *SpaceSvc) List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Get(ctx context.Context, id string) (*spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Update(ctx context.Context, id string, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Delete(ctx context.Context, id string) error {
	// Implementation logic here
	return nil
}
