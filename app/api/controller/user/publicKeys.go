package user

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// CreatePublicKey godoc
// @Summary 创建用户公钥
// @Description 创建用户公钥
// @Tags 用户
// @Accept json
// @Produce json
// @Param input body publickey.CreatePublicKeyInput true "创建用户公钥参数"
// @Success 200 {object} resp.ObjectResponse "创建用户公钥成功"
// @Failure 400 {object} resp.ObjectResponse "创建用户公钥失败"
// @Router /user/public-keys [post]
func (c *Controller) Create(ctx *gin.Context) {
	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.User{
			ID:    1,
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}

	var params publickey.CreatePublicKeyInput
	var err error
	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.AuthPubKeyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := params.CheckIdentifier(); err != nil {
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
