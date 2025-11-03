package space

import "time"

type CreateSpaceRequest struct {
	Name       string `json:"name" binding:"required,max=200" example:"测试空间01"`
	Owner      string `json:"owner" binding:"required"`
	Permission string `json:"acl" binding:"required,oneof=open private" example:"open"`
	Desc       string `json:"desc"`
	CreatedBy  string `json:"createdBy" binding:"required"`
}

type UpdateSpaceRequest struct {
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Permission string `json:"acl"`
	Desc       string `json:"desc"`
	EditedBy   string `json:"editedBy" binding:"required"`
}

type SpaceResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Owner       string     `json:"owner"`
	Permission  string     `json:"acl"`
	Desc        string     `json:"desc"`
	CreatedBy   string     `json:"createdBy"`
	CreatedDate *time.Time `json:"createdDate"`
	EditedBy    string     `json:"editedBy"`
	EditedDate  *time.Time `json:"editedDate"`
}

type ListSpaceFilter struct {
	Name       string `form:"name" binding:"omitempty"` // 非必填，omitempty 允许空值
	Permission string `json:"name" form:"permission" binding:"omitempty,oneof=open private"` // 枚举校验：仅允许 open/private
	Owner      string `form:"owner" binding:"omitempty"`

	// 分页参数（对应 pagination.PageParam）
	Page     int `form:"page" binding:"omitempty,min=1"` // 最小1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"` // 1-100
}

type UndeleteSpaceRequest struct {
	IDs      []uint16 `json:"ids" binding:"required,min=1,dive,gt=0" example:"1,2,3"`
	EditedBy string   `json:"editedBy" binding:"required" example:"admin"`
}
