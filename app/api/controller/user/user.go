package user

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
)

// Controller 处理与提交相关的 HTTP 请求
type Controller struct {
	userSvc      userSvc.Interface
	publicKeySvc publickey.Interface
}

// NewUserController 创建提交处理器实例
func NewUserController(userService userSvc.Interface, publicKeyService publickey.Interface) *Controller {
	return &Controller{
		userSvc:      userService,
		publicKeySvc: publicKeyService,
	}
}
