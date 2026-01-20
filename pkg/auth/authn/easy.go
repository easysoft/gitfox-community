package authn

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/cryptor"
	"github.com/gin-gonic/gin"
)

const (
	HeaderApp      = "App"
	HeaderOperator = "Operator"
)

var _ Authenticator = (*EasyAuthenticator)(nil)

// EasyAuthenticator uses App, Operator and Authorization headers to authenticate the caller.
// EasyAuthenticator 使用 App、Operator 和 Authorization 请求头来验证调用者。
//
// 鉴权机制：
// 1. 前端计算：Authorization = md5(app + operator + token)
// 2. 后端验证：从请求头获取 app、operator，从数据库获取 token，计算 md5 并与 Authorization 对比
//
// 请求示例：
// curl --location --request DELETE 'http://example.com/api/v2/repos/2' \
// --header 'Authorization: 7f18226c955beb483985aed5c581d0ba' \
// --header 'App: zentao' \
// --header 'Operator: zhangsan'
type EasyAuthenticator struct {
	userStore  user.Interface
	tokenStore token.Interface
}

// NewEasyAuthenticator creates a new EasyAuthenticator.
// NewEasyAuthenticator 创建一个新的 EasyAuthenticator。
func NewEasyAuthenticator(userStore user.Interface, tokenStore token.Interface) *EasyAuthenticator {
	return &EasyAuthenticator{
		userStore:  userStore,
		tokenStore: tokenStore,
	}
}

// Authenticate validates the request using Easy authentication mechanism.
// Authenticate 使用 Easy 鉴权机制验证请求。
func (a *EasyAuthenticator) Authenticate(r *gin.Context) (*auth.Session, error) {
	logger := logger.GetLogger(r)
	logger.Debug("[authn.easy]trigger EasyAuthenticator")

	var authorization, app, operator string
	// 1. 从请求头获取必要的认证信息
	authorization, _ = req.GetHeader(r, HeaderAuthorization)
	if authorization == "" {
		authorization, _ = req.QueryParam(r, HeaderAuthorization)
	}
	if authorization == "" {
		logger.Debug("[authn.easy]missing Authorization header")
		return nil, constant.ErrAuthNoData
	}

	app, _ = req.GetHeader(r, HeaderApp)
	if app == "" {
		app, _ = req.QueryParam(r, HeaderApp)
	}
	if app == "" {
		logger.Debug("[authn.easy]missing App header")
		return nil, constant.ErrAuthNoData
	}

	operator, _ = req.GetHeader(r, HeaderOperator)
	if operator == "" {
		operator, _ = req.QueryParam(r, HeaderOperator)
	}

	if operator == "" {
		logger.Debug("[authn.easy]missing Operator header")
		return nil, constant.ErrAuthNoData
	}

	// 2. 根据 operator（用户账号）查询用户信息
	user, err := a.userStore.FindByAccount(r.Request.Context(), operator)
	if err != nil {
		if errors.Is(err, errors.ErrResourceNotFound) {
			logger.Debugf("[authn.easy]user %s not found", operator)
			return nil, constant.ErrUnauthorized
		}
		logger.Debugf("[authn.easy]failed to find user %s: %v", operator, err)
		return nil, constant.ErrUnauthorized
	}

	if user == nil {
		logger.Debugf("[authn.easy]user %s is nil", operator)
		return nil, constant.ErrUnauthorized
	}

	// 3. 从数据库获取用户的 token
	// 使用 app 作为 identifier 来查询 token
	entry, err := a.tokenStore.GetDevopsToken(r.Request.Context())
	if err != nil {
		logger.Debugf("[authn.easy]failed to get token for user %s: %v", operator, err)
		return nil, constant.ErrUnauthorized
	}

	if entry == nil || entry.Key == "" {
		logger.Debugf("[authn.easy]token is empty for user %s", operator)
		return nil, constant.ErrUnauthorized
	}

	// 4. 计算期望的 Authorization 值：md5(app + operator + token)
	expectedAuth := cryptor.Md5String(app + operator + entry.Key)

	// 5. 比对 Authorization 值（不区分大小写）
	if !strings.EqualFold(authorization, expectedAuth) {
		logger.Debugf("[authn.easy]invalid authorization for user %s, expected: %s, got: %s",
			operator, expectedAuth, authorization)
		return nil, constant.ErrUnauthorized
	}

	var isAdmin bool
	if app == HeaderAppAdmin {
		isAdmin = true
	}

	logger.Debugf("[authn.easy]user %s authenticated successfully", operator)
	return &auth.Session{
		Principal: &principal.Principal{
			App:         app,
			DisplayName: app,
			Admin:       isAdmin,
			Email:       HeaderAdminEmail,
		},
		User: user,
	}, nil
}

const (
	HeaderAppAdmin   = "zentao"
	HeaderAdminEmail = "devops@chandao.com"
)
