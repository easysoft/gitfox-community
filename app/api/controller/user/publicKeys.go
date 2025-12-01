package user

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// CreatePublicKey godoc
// @Summary 添加用户公钥
// @Description 添加用户公钥
// @Tags 用户
// @Accept json
// @Produce json
// @Param input body publickey.AddPublicKeyInput true "创建用户公钥参数"
// @Success 200 {object} resp.ObjectResponse "创建用户公钥成功"
// @Failure 400 {object} resp.ObjectResponse "创建用户公钥失败"
// @Router /user/keys [post]
func (c *Controller) AddKey(ctx *gin.Context) {
	var params publickey.AddPublicKeyInput
	var err error
	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyAddParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthPubKeyAddError)
		resp.BadRequest(ctx, err)
		return
	}

	if err = params.CheckIdentifier(); err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyIdentifierError)
		resp.BadRequest(ctx, err)
		return
	}

	params.Content = strings.TrimSpace(params.Content)
	if params.Content == "" {
		err = errors.WithI18nKey(errors.New("public key not provided"), i18n.AuthPubKeyAddError)
		resp.BadRequest(ctx, err)
		return
	}

	err = c.publicKeySvc.CreatePublicKey(ctx, session, &params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyAddError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthPubKeyAddSuccess, nil)

}

// ListKeys godoc
// @Summary 获取用户公钥列表
// @Description 获取用户公钥列表
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} resp.ObjectResponse "获取用户公钥列表成功"
// @Failure 400 {object} resp.ObjectResponse "获取用户公钥列表失败"
// @Router /user/keys [get]
func (c *Controller) ListKeys(ctx *gin.Context) {
	var params req.ListFilter
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}
	if err = ctx.ShouldBindQuery(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyListParamError)
		resp.BadRequest(ctx, err)
		return
	}

	keys, err := c.publicKeySvc.ListPublicKeys(ctx, session, params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyListError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.ListSuccess(ctx, keys.Data,
		resp.ListArgs{
			Total:    keys.Total,
			Page:     int64(keys.Page),
			PageSize: int64(keys.PageSize),
		}, i18n.AuthPubKeyListSuccess, nil)
}
