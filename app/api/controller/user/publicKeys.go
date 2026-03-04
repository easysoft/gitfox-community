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

// AddKey 添加用户公钥
// @Summary 添加用户公钥
// @Description 为当前登录用户添加一条 SSH 公钥
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param input body publickey.AddPublicKeyInput true "添加用户公钥参数"
// @Success 200 {object} resp.ObjectResponse "添加成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
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
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err = params.Check(); err != nil {
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

// ListKeys 获取用户公钥列表
// @Summary 获取用户公钥列表
// @Description 查询当前登录用户的 SSH 公钥列表（按创建时间倒序）
// @Tags users
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ListResponse{data=[]publickey.PublicKey,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /user/keys [get]
func (c *Controller) ListKeys(ctx *gin.Context) {
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	keys, _, err := c.publicKeySvc.ListWithTotal(ctx, session)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyListError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthPubKeyListSuccess, keys)
}

// GetKey 获取用户公钥详情
// @Summary 获取用户公钥详情
// @Description 按ID获取当前登录用户的 SSH 公钥详情
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param id path int true "公钥ID"
// @Success 200 {object} resp.ObjectResponse{data=publickey.PublicKey} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 404 {object} resp.ObjectResponse "公钥不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /user/keys/{id} [get]
func (c *Controller) GetKey(ctx *gin.Context) {
	var err error

	id, err := req.PathParamAsPositiveInt64(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyListParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	key, err := c.publicKeySvc.GetPublicKeyByID(ctx, session, id)
	if err != nil {
		if errors.IsNotFound(err) {
			resp.NotFound(ctx, err)
			return
		}
		err = errors.WithI18nKey(err, i18n.AuthPubKeyListError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthPubKeyListSuccess, key)
}

// UpdateKey 修改用户公钥
// @Summary 修改用户公钥
// @Description 按ID修改当前登录用户的 SSH 公钥，仅允许修改标识符和公钥内容
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param id path int true "公钥ID"
// @Param input body publickey.UpdatePublicKeyInput true "修改用户公钥参数"
// @Success 200 {object} resp.ObjectResponse "修改成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 404 {object} resp.ObjectResponse "公钥不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /user/keys/{id} [put]
func (c *Controller) UpdateKey(ctx *gin.Context) {
	var params publickey.UpdatePublicKeyInput
	var err error

	id, err := req.PathParamAsPositiveInt64(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyUpdateParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyUpdateParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err = params.Check(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	if params.Content != "" {
		params.Content = strings.TrimSpace(params.Content)
	}

	err = c.publicKeySvc.UpdatePublicKey(ctx, session, id, &params)
	if err != nil {
		if errors.IsNotFound(err) {
			resp.NotFound(ctx, err)
			return
		}
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthPubKeyUpdateSuccess, nil)
}

// DeleteKey 删除用户公钥
// @Summary 删除用户公钥
// @Description 按ID硬删除当前登录用户的 SSH 公钥
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param id path int true "公钥ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 404 {object} resp.ObjectResponse "公钥不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /user/keys/{id} [delete]
func (c *Controller) DeleteKey(ctx *gin.Context) {
	var err error

	id, err := req.PathParamAsPositiveInt64(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyDeleteParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	err = c.publicKeySvc.Delete(ctx, session, id)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyDeleteError)
		if errors.IsNotFound(err) {
			resp.NotFound(ctx, err)
			return
		}
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthPubKeyDeleteSuccess, nil)
}
