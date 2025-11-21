package authn

import (
	"crypto/md5"
	"encoding/hex"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

var _ Authenticator = (*BasicAuthenticator)(nil)

// BasicAuthenticator uses the provided JWT to authenticate the caller.
// BasicAuthenticator 使用提供的JWT来验证调用者。
type BasicAuthenticator struct {
	principalStore principal.Interface
}

// NewBasicAuthenticator creates a new BasicAuthenticator.
// NewBasicAuthenticator 创建一个新的BasicAuthenticator。
func NewBasicAuthenticator(principalStore principal.Interface) *BasicAuthenticator {
	return &BasicAuthenticator{
		principalStore: principalStore,
	}
}

func (a *BasicAuthenticator) Authenticate(r *gin.Context) (*auth.Session, error) {

	logger := logger.GetLogger(r)
	logger.Debug("[authn.basic]trigger BasicAuthenticator")

	userName, passwd, ok := r.Request.BasicAuth()
	if !ok {
		return nil, constant.ErrAuthNoData
	}

	user, err := a.principalStore.FindUserByUID(r, userName)
	if errors.Is(err, errors.ErrResourceNotFound) {
		return nil, constant.ErrUnauthorized
	}

	hash := md5.Sum([]byte(passwd))
	hashedPassword := hex.EncodeToString(hash[:])
	if hashedPassword != user.Password {
		logger.Debugf("[authn.basic]user %s has invalid zentao password", user.UID)
		return nil, constant.ErrUnauthorized
	}

	logger.Debugf("[authn.basic]user %s authenticated", user.UID)
	return &auth.Session{
		User: user,
	}, nil
}
