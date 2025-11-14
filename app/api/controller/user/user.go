package user

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Controller 处理与提交相关的 HTTP 请求
type Controller struct {
	userSvc userSvc.Interface
}

// NewController 创建提交处理器实例
func NewController(userService userSvc.Interface) *Controller {
	return &Controller{
		userSvc: userService,
	}
}

func (c *Controller) Login(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params auth.LoginInput
	var err error
	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	user, err := c.userSvc.FindUserByUID(ctx, params.LoginIdentifier)
	if err != nil && errors.GetI18nKey(err) == i18n.DbNotfound {
		user, err = c.userSvc.FindUserByEmail(ctx, params.LoginIdentifier)
	}

	if err != nil {
		logger.Errorf("FindUserByUID or FindUserByEmail failed: %v, loginIdentifier: %s", err, params.LoginIdentifier)
		err = errors.WithI18nKey(err, i18n.AuthUserNotFound)
		resp.BadRequest(ctx, err)
		return
	}

	logger.Infof("user found: %v", user)

}
