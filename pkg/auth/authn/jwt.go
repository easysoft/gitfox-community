package authn

import (
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	userModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"github.com/gin-gonic/gin"

	gojwt "github.com/golang-jwt/jwt"
)

const HeaderAuthorization = "Authorization"
const CookieName = "token"

var _ Authenticator = (*BasicAuthenticator)(nil)

type JWTAuthenticator struct {
	principalStore principal.Interface
	TokenStore     token.Interface
}

func NewJWTAuthenticator(principalStore principal.Interface, tokenStore token.Interface) *JWTAuthenticator {
	return &JWTAuthenticator{
		principalStore: principalStore,
		TokenStore:     tokenStore,
	}
}

func (a *JWTAuthenticator) Authenticate(r *gin.Context) (*auth.Session, error) {
	logger := logger.GetLogger(r)
	token := extractToken(r)
	if token == "" {
		return nil, constant.ErrAuthNoData
	}

	var principal *userModel.User
	var err error
	claims := &Claims{}
	parsed, err := gojwt.ParseWithClaims(
		token, claims, func(_ *gojwt.Token) (interface{}, error) {
			principal, err = a.principalStore.FindUserByID(r.Request.Context(), claims.PrincipalID)
			if err != nil {
				return nil, fmt.Errorf("failed to get principal for token: %w", err)
			}
			if principal.Blocked {
				return nil, fmt.Errorf("user is blocked")
			}
			return []byte(principal.Salt), nil
		},
	)

	if err != nil {
		return nil, errors.Newf("parsing of JWT claims failed: %w", err)
	}

	if !parsed.Valid {
		return nil, errors.New("parsed JWT token is invalid")
	}

	if _, ok := parsed.Method.(*gojwt.SigningMethodHMAC); !ok {
		return nil, errors.New("invalid HMAC signature for JWT")
	}

	// 试图从query中获取sudo用户ID
	sudoUserID, err := req.QueryParamAsPositiveInt64OrDefault(r, req.QueryParamSudoUserID, 0)
	if err != nil {
		logger.Warnf("failed to get sudo user ID from query: %v", err)
	}

	// 试图从sudo用户ID获取用户
	user, err := a.principalStore.FindUserByID(r.Request.Context(), sudoUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get principal for token: %w", err)
	}

	return &auth.Session{
		Principal: *principal,
		User:      user,
	}, nil
}

func extractToken(r *gin.Context) string {

	// Check query param first (as that's most immediately visible to caller)
	// 检查查询参数是否包含访问令牌
	queryToken := req.QueryParamOrDefault(r, req.QueryParamAccessToken, "")
	if queryToken != "" {
		return queryToken
	}

	// check authorization header next
	// 接下来检查授权标头
	headerToken := r.GetHeader(HeaderAuthorization)
	switch {
	// in case of git push / pull it would be basic auth and token is in password
	// 在 git push / pull 的情况下，它将是基本身份验证，令牌在密码中
	case strings.HasPrefix(headerToken, "Basic "):
		// return pwd either way - if it's invalid pwd is empty string which we'd return anyway
		// 无论是否有效，都返回密码 - 如果无效，密码为空字符串，我们无论如何都将返回
		_, pwd, _ := r.Request.BasicAuth()
		return pwd
	// strip bearer prefix if present
	// 如果存在，删除 Bearer 前缀
	case strings.HasPrefix(headerToken, "Bearer "):
		return headerToken[7:]
	// otherwise use value as is
	// 否则直接使用值
	case headerToken != "":
		return headerToken
	}

	// check cookies last (as that's least visible to caller)
	// 最后检查 Cookie（因为对调用者来说，它是最不可见的）
	if cookieToken, ok := req.GetCookie(r, CookieName); ok {
		return cookieToken
	}

	// no token found
	// 没有找到令牌
	return ""
}
