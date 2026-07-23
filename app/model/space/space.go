package space

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	common "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type CreateSpaceRequest struct {
	Name      string `json:"name" binding:"required,max=200" example:"测试空间01"`
	ACL       string `json:"acl" binding:"required,oneof=open private" example:"open"`
	Code      string `json:"code" binding:"required,max=50" example:"test_space_01"`
	Auth      string `json:"auth" binding:"required,oneof=extend reset" example:"extend"`
	Desc      string `json:"desc"`
	CreatedBy string `json:"createdBy" binding:"required"`
}

func (req *CreateSpaceRequest) Validate() error {
	// 校验Code字符规则（符合Git仓库命名规范）
	if !common.IsValidCode(req.Code) {
		return errors.WithI18nKey(nil, i18n.SpaceCodeInvalid)
	}

	if len(req.Code) > 50 {
		return errors.WithI18nKey(nil, i18n.SpaceCodeTooLong)
	}

	codeLower := strings.ToLower(req.Code)
	if strings.HasSuffix(codeLower, ".git") {
		return errors.WithI18nKey(nil, i18n.SpaceCodeInvalid)
	}

	if len(req.Desc) > 500 {
		return errors.WithI18nKey(nil, i18n.SpaceDescTooLong)
	}

	return nil
}

type UpdateSpaceRequest struct {
	Name     string `json:"name"`
	Auth     string `json:"auth" binding:"required,oneof=extend reset" example:"extend"`
	ACL      string `json:"acl" binding:"omitempty,oneof=open private"`
	Desc     string `json:"desc,omitempty"`
	EditedBy string `json:"editedBy" binding:"required"`
}

type SpaceResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Code        string          `json:"code"`
	Auth        string          `json:"auth"`
	ACL         string          `json:"acl"`
	Desc        string          `json:"desc"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

type ListSpaceFilter struct {
	Name    string `form:"name" binding:"omitempty"`                              // 非必填，omitempty 允许空值
	Code    string `form:"code" binding:"omitempty"`                              // 非必填，omitempty 允许空值
	ACL     string `json:"acl" form:"acl" binding:"omitempty,oneof=open private"` // 枚举校验：仅允许 open/private
	IDs     string `json:"ids" form:"ids" binding:"omitempty,gt=0"`               //gt=0 要求元素大于0
	Deleted bool   `form:"deleted" binding:"omitempty"`                           // 是否包含已删除空间
	IsOpen  bool   `form:"isOpen" binding:"omitempty"`                            // 是否公开空间

	// 分页参数（对应 pagination.PageParam）
	Page     int `form:"page" binding:"omitempty,min=1"`              // 最小1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=2000"` // 1-2000
}

type UndeleteSpaceRequest struct {
	IDs      []uint16 `json:"ids" binding:"required,min=1,dive,gt=0" example:"1,2,3"`
	EditedBy string   `json:"editedBy" binding:"required" example:"admin"`
}
