package user

import (
	"net/http"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	userSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	authPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"github.com/gin-gonic/gin"
)

// Controller 处理与提交相关的 HTTP 请求
type Controller struct {
	userSvc      userSvc.Interface
	publicKeySvc publickey.Interface
}

// NewController 创建提交处理器实例
func NewController(userService userSvc.Interface, publicKeyService publickey.Interface) *Controller {
	return &Controller{
		userSvc:      userService,
		publicKeySvc: publicKeyService,
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

	// 查找用户
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

	// 比较密码
	if err := c.userSvc.CompareHashAndPassword(ctx, []byte(user.Password), []byte(params.Password)); err != nil {
		logger.Debugf("invalid password, user_uid: %s", user.UID)
		err = errors.WithI18nKey(err, i18n.AuthUserPassError)
		resp.BadRequest(ctx, err)
		return
	}

	// 下发token
	token, err := c.userSvc.IssueToken(ctx, user)
	if err != nil {
		logger.Errorf("IssueUserSession failed: %v, user_uid: %s", err, user.UID)
		err = errors.WithI18nKey(err, i18n.AuthUserIssueTokenError)
		resp.BadRequest(ctx, err)
		return
	}

	includeTokenCookie(ctx, token)

	// 管理员才会立马下发pat
	if !user.IsAdmin {
		resp.OK(ctx, i18n.AuthAuthUserLoginSuccess, auth.LoginResp{
			Token: token,
			PAT:   nil,
		})
		return
	}

	// 下发pat
	pat, err := c.userSvc.IssuePAT(ctx, user, authPkg.AdminPATIdentifier, nil)
	if err != nil {
		logger.Errorf("IssuePAT failed: %v, user_uid: %s", err, user.UID)
		err = errors.WithI18nKey(err, i18n.AuthUserIssuePatError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.AuthAuthUserLoginSuccess, auth.LoginResp{
		Token: token,
		PAT:   pat,
	})

}

func includeTokenCookie(r *gin.Context, tokenResponse *auth.TokenResponse) {
	cookie := newEmptyTokenCookie(r.Request, authPkg.CookieName)
	cookie.Value = tokenResponse.AccessToken
	if tokenResponse.Token.ExpiresAt != nil {
		cookie.Expires = time.UnixMilli(*tokenResponse.Token.ExpiresAt)
	}

	r.Writer.Header().Add("Set-Cookie", cookie.String())
}

func newEmptyTokenCookie(r *http.Request, cookieName string) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Path:     "/",
		Domain:   r.URL.Hostname(),
		Secure:   r.URL.Scheme == "https",
	}
}
