package user

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

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
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
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
