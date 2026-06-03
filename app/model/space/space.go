package space

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type CreateSpaceRequest struct {
	Name       string `json:"name" binding:"required,max=200" example:"测试空间01"`
	Permission string `json:"acl" binding:"required,oneof=open private" example:"open"`
	Auth       string `json:"auth" binding:"required,oneof=extend reset" example:"extend"`
	Desc       string `json:"desc"`
	CreatedBy  string `json:"createdBy" binding:"required"`
}

type UpdateSpaceRequest struct {
	Name       string `json:"name"`
	Auth       string `json:"auth" binding:"required,oneof=extend reset" example:"extend"`
	Permission string `json:"acl" binding:"omitempty,oneof=open private"`
	Desc       string `json:"desc,omitempty"`
	EditedBy   string `json:"editedBy" binding:"required"`
}

type SpaceResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Auth        string          `json:"auth"`
	Permission  string          `json:"acl"`
	Desc        string          `json:"desc"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

type ListSpaceFilter struct {
	Name       string `form:"name" binding:"omitempty"`                                      // 非必填，omitempty 允许空值
	Permission string `json:"name" form:"permission" binding:"omitempty,oneof=open private"` // 枚举校验：仅允许 open/private
	IDs        string `json:"ids" form:"ids" binding:"omitempty,gt=0"`                       //gt=0 要求元素大于0
	Deleted    bool   `form:"deleted" binding:"omitempty"`                                   // 是否包含已删除空间
	IsOpen     bool   `form:"isOpen" binding:"omitempty"`                                    // 是否公开空间

	// 分页参数（对应 pagination.PageParam）
	Page     int `form:"page" binding:"omitempty,min=1"`             // 最小1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=2000"` // 1-2000
}

type UndeleteSpaceRequest struct {
	IDs      []uint16 `json:"ids" binding:"required,min=1,dive,gt=0" example:"1,2,3"`
	EditedBy string   `json:"editedBy" binding:"required" example:"admin"`
}
