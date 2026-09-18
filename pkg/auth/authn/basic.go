package authn

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
	"github.com/easysoft/gitfox-community/app/store/user"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

var _ Authenticator = (*BasicAuthenticator)(nil)

// BasicAuthenticator uses the provided JWT to authenticate the caller.
// BasicAuthenticator 使用提供的JWT来验证调用者。
type BasicAuthenticator struct {
	userStore    user.Interface
	pipelineAuth config.PipelineAuthConfig
}

// NewBasicAuthenticator creates a new BasicAuthenticator.
// NewBasicAuthenticator 创建一个新的BasicAuthenticator。
func NewBasicAuthenticator(userStore user.Interface, pipelineAuth config.PipelineAuthConfig) *BasicAuthenticator {
	return &BasicAuthenticator{
		userStore:    userStore,
		pipelineAuth: pipelineAuth,
	}
}

// 请求示例
// Authorization: Basic base64(username:password)
// Authorization: Basic YWRtaW46MTIzNDU2
func (a *BasicAuthenticator) Authenticate(r *gin.Context) (*auth.Session, error) {

	logger := logger.GetLogger(r)
	logger.Debug("[authn.basic]trigger BasicAuthenticator")

	userName, passwd, ok := r.Request.BasicAuth()
	if !ok {
		return nil, constant.ErrAuthNoData
	}

	if userName == a.pipelineAuth.UserName {
		if passwd != a.pipelineAuth.Password {
			return nil, constant.ErrUnauthorized
		}
		userInfo := principal.User{
			Account: a.pipelineAuth.UserName,
			Email:   HeaderAdminEmail,
		}
		return &auth.Session{User: &userInfo}, nil
	}

	user, err := a.userStore.FindByAccount(r, userName)
	if errors.Is(err, errors.ErrResourceNotFound) {
		return nil, constant.ErrUnauthorized
	}

	if user == nil {
		return nil, constant.ErrUnauthorized
	}

	hash := md5.Sum([]byte(passwd))
	hashedPassword := hex.EncodeToString(hash[:])
	if hashedPassword != user.Password {
		logger.Debugf("[authn.basic]user %s has invalid zentao password", user.Account)
		return nil, constant.ErrUnauthorized
	}

	logger.Debugf("[authn.basic]user %s authenticated", user.Account)
	return &auth.Session{
		User: user,
	}, nil
}
