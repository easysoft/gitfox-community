package user

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/service/publickey"
	userSvc "github.com/easysoft/gitfox-community/app/service/user"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ensure authModel is used for swagger generation
var _ *authModel.Session

// Controller 处理与用户相关的 HTTP 请求
type Controller struct {
	userSvc      userSvc.Interface
	publicKeySvc publickey.Interface
}

// NewUserController 创建用户处理器实例
func NewUserController(userService userSvc.Interface, publicKeyService publickey.Interface) *Controller {
	return &Controller{
		userSvc:      userService,
		publicKeySvc: publicKeyService,
	}
}

// Current 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的基础信息，包括账号、邮箱、显示名称、管理员状态和封禁状态
// @Tags users
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=authModel.Session} "查询成功"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 404 {object} resp.ObjectResponse "用户不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /fit/user [get]
func (c *Controller) Current(ctx *gin.Context) {
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	resp.OK(ctx, i18n.AuthUserProfileQuerySuccess, session)
}
