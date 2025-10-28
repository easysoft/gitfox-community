package space

import "time"

type CreateSpaceRequest struct {
	Name       string `json:"name" binding:"required,max=200" example:"测试空间01"`
	Owner      string `json:"owner" binding:"required"`
	Permission string `json:"acl" binding:"required,oneof=open private" example:"open"`
	Desc       string `json:"desc"`
	CreatedBy  string `json:"createdBy" binding:"required"`
}

type SpaceResponse struct {
	ID          uint16     `json:"id"`
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
	Name       string
	Permission string `json:"acl"`
	Owner      string
	Page       int
	PageSize   int
}

type UpdateSpaceRequest struct {
	Name       string
	Owner      string
	Permission string `json:"acl"`
	Desc       string
	EditedBy   string
}
