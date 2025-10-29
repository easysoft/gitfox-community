package space

import (
	"context"

	spaceDao "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/dao/space"
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
	in := spaceDao.Space{
		Name:       req.Name,
		Desc:       req.Desc,
		Owner:      req.Owner,
		Permission: req.Permission,
		CreatedBy:  req.CreatedBy,
		EditedBy:   req.CreatedBy,
	}

	out, err := s.spaceStore.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	return &spaceModel.SpaceResponse{
		ID:          out.ID,
		Name:        out.Name,
		Owner:       out.Owner,
		Permission:  out.Permission,
		Desc:        out.Desc,
		CreatedBy:   out.CreatedBy,
		CreatedDate: out.CreatedDate,
		EditedBy:    out.EditedBy,
		EditedDate:  out.EditedDate,
	}, nil
}

func (s *SpaceSvc) List(ctx context.Context, filter spaceModel.ListSpaceFilter) ([]spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Get(ctx context.Context, id uint16) (*spaceModel.SpaceResponse, error) {
	// Implementation logic here
	return nil, nil
}

func (s *SpaceSvc) Update(ctx context.Context, id uint16, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {

	out, err := s.spaceStore.Update(ctx, id, toUpdateMap(req))
	if err != nil {
		return nil, err
	}

	return &spaceModel.SpaceResponse{
		ID:          out.ID,
		Name:        out.Name,
		Owner:       out.Owner,
		Permission:  out.Permission,
		Desc:        out.Desc,
		CreatedBy:   out.CreatedBy,
		CreatedDate: out.CreatedDate,
		EditedBy:    out.EditedBy,
		EditedDate:  out.EditedDate,
	}, nil
}

func (s *SpaceSvc) Delete(ctx context.Context, id uint16) error {
	// Implementation logic here
	return nil
}

func toUpdateMap(req spaceModel.UpdateSpaceRequest) map[string]interface{} {
	m := map[string]interface{}{
		"`editedBy`": req.EditedBy, // 必填，始终包含
	}
	if req.Name != "" {
		m["`name`"] = req.Name
	}
	if req.Owner != "" {
		m["`owner`"] = req.Owner
	}
	if req.Permission != "" {
		m["`acl`"] = req.Permission
	}
	if req.Desc != "" {
		m["`desc`"] = req.Desc
	}
	return m
}
