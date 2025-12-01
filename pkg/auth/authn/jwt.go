package authn

import (
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	userModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/user"
	"github.com/gin-gonic/gin"

	gojwt "github.com/golang-jwt/jwt"
)

const HeaderAuthorization = "Authorization"
const CookieName = "token"
const AdminPATIdentifier = "adminPat"

var _ Authenticator = (*BasicAuthenticator)(nil)

type JWTAuthenticator struct {
	userStore  user.Interface
	TokenStore token.Interface
}

func NewJWTAuthenticator(userStore user.Interface, tokenStore token.Interface) *JWTAuthenticator {
	return &JWTAuthenticator{
		userStore:  userStore,
		TokenStore: tokenStore,
	}
}

func (a *JWTAuthenticator) Authenticate(r *gin.Context) (*auth.Session, error) {
	token := extractToken(r)
	if token == "" {
		return nil, constant.ErrAuthNoData
	}

	principal := &userModel.Principal{}
	var user *userModel.User
	var err error
	claims := &Claims{}
	parsed, err := gojwt.ParseWithClaims(
		token, claims, func(_ *gojwt.Token) (interface{}, error) {
			if claims.App == "" { // 用户令牌
				temp, err := a.userStore.FindUserByAccount(r.Request.Context(), claims.Account)
				if err != nil {
					return nil, fmt.Errorf("failed to get principal for token: %w", err)
				}
				user = temp
			}
			principal.Admin = claims.Admin
			principal.App = claims.App
			principal.DisplayName = claims.Account
			return []byte(KeysToSecret(claims.Account, claims.Identifier)), nil
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "parsing of JWT claims failed")
	}

	if !parsed.Valid {
		return nil, errors.New("parsed JWT token is invalid")
	}

	if _, ok := parsed.Method.(*gojwt.SigningMethodHMAC); !ok {
		return nil, errors.New("invalid HMAC signature for JWT")
	}

	if user == nil {
		operation := req.GetHeaderOrDefault(r, req.QueryParamOperation, "")
		if operation != "" {
			user, err = a.userStore.FindUserByAccount(r.Request.Context(), operation)
			if err != nil {
				return nil, fmt.Errorf("failed to get user for token: %w", err)
			}
		}
	}

	return &auth.Session{
		Principal: principal,
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

	// no token found
	// 没有找到令牌
	return ""
}

func GenerateJWTToken(claims *Claims, secret []byte) (string, error) {
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
