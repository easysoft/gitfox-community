package authn

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	userModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/token"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockUserStore 模拟用户存储
type mockUserStore struct {
	users map[string]*userModel.User
}

func (m *mockUserStore) FindUserByID(ctx context.Context, id int64) (*userModel.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.ErrResourceNotFound
}

func (m *mockUserStore) FindUserByAccount(ctx context.Context, account string) (*userModel.User, error) {
	user, ok := m.users[account]
	if !ok {
		return nil, errors.ErrResourceNotFound
	}
	return user, nil
}

func (m *mockUserStore) FindUserByEmail(ctx context.Context, email string) (*userModel.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.ErrResourceNotFound
}

// mockTokenStore 模拟 token 存储
type mockTokenStore struct {
	tokens map[string]*token.Token // key: account_identifier
}

func (m *mockTokenStore) GetByID(ctx context.Context, id int64) (*token.Token, error) {
	return nil, errors.ErrResourceNotFound
}

func (m *mockTokenStore) GetByIdentifier(ctx context.Context, account string, identifier string) (*token.Token, error) {
	key := account + "_" + identifier
	tok, ok := m.tokens[key]
	if !ok {
		return nil, errors.ErrResourceNotFound
	}
	return tok, nil
}

func (m *mockTokenStore) UpdateToken(ctx context.Context, id int64, tokenStr string) error {
	return nil
}

func (m *mockTokenStore) CreateUserPAT(ctx context.Context, user *userModel.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	return nil, nil
}

func (m *mockTokenStore) CreateAppPAT(ctx context.Context, user *userModel.User, tokenIdentifier string, lifetime *time.Duration) (*auth.Token, error) {
	return nil, nil
}

func TestEasyAuthenticator_Authenticate_Success(t *testing.T) {
	// 准备测试数据
	account := "zhangsan"
	app := "zentao"
	tokenValue := "xxxx"

	// 计算期望的 Authorization 值
	expectedAuth := calculateMD5(app + account + tokenValue)

	// 创建 mock stores
	userStore := &mockUserStore{
		users: map[string]*userModel.User{
			account: {
				ID:      1,
				Account: account,
				Email:   "zhangsan@example.com",
			},
		},
	}

	tokenStore := &mockTokenStore{
		tokens: map[string]*token.Token{
			account + "_" + app: {
				ID:         1,
				Account:    account,
				Identifier: app,
				Token:      tokenValue,
			},
		},
	}

	// 创建认证器
	authenticator := NewEasyAuthenticator(userStore, tokenStore)

	// 创建 Gin 上下文
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置请求头
	req, _ := http.NewRequest("DELETE", "/api/v2/repos/2", nil)
	req.Header.Set("Authorization", expectedAuth)
	req.Header.Set("App", app)
	req.Header.Set("Operator", account)
	c.Request = req

	// 执行认证
	session, err := authenticator.Authenticate(c)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotNil(t, session.User)
	assert.Equal(t, account, session.User.Account)
	assert.Equal(t, int64(1), session.User.ID)
}

func TestEasyAuthenticator_Authenticate_MissingHeaders(t *testing.T) {
	userStore := &mockUserStore{users: map[string]*userModel.User{}}
	tokenStore := &mockTokenStore{tokens: map[string]*token.Token{}}
	authenticator := NewEasyAuthenticator(userStore, tokenStore)

	tests := []struct {
		name          string
		authorization string
		app           string
		operator      string
	}{
		{"missing authorization", "", "zentao", "zhangsan"},
		{"missing app", "abc123", "", "zhangsan"},
		{"missing operator", "abc123", "zentao", ""},
		{"missing all", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("DELETE", "/api/v2/repos/2", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			if tt.app != "" {
				req.Header.Set("App", tt.app)
			}
			if tt.operator != "" {
				req.Header.Set("Operator", tt.operator)
			}
			c.Request = req

			session, err := authenticator.Authenticate(c)

			assert.Error(t, err)
			assert.Nil(t, session)
			assert.Equal(t, constant.ErrAuthNoData, err)
		})
	}
}

func TestEasyAuthenticator_Authenticate_UserNotFound(t *testing.T) {
	userStore := &mockUserStore{users: map[string]*userModel.User{}}
	tokenStore := &mockTokenStore{tokens: map[string]*token.Token{}}
	authenticator := NewEasyAuthenticator(userStore, tokenStore)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("DELETE", "/api/v2/repos/2", nil)
	req.Header.Set("Authorization", "someauth")
	req.Header.Set("App", "zentao")
	req.Header.Set("Operator", "nonexistent")
	c.Request = req

	session, err := authenticator.Authenticate(c)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, constant.ErrUnauthorized, err)
}

func TestEasyAuthenticator_Authenticate_InvalidAuthorization(t *testing.T) {
	account := "zhangsan"
	app := "zentao"
	tokenValue := "xxxx"

	userStore := &mockUserStore{
		users: map[string]*userModel.User{
			account: {
				ID:      1,
				Account: account,
				Email:   "zhangsan@example.com",
			},
		},
	}

	tokenStore := &mockTokenStore{
		tokens: map[string]*token.Token{
			account + "_" + app: {
				ID:         1,
				Account:    account,
				Identifier: app,
				Token:      tokenValue,
			},
		},
	}

	authenticator := NewEasyAuthenticator(userStore, tokenStore)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 使用错误的 Authorization 值
	req, _ := http.NewRequest("DELETE", "/api/v2/repos/2", nil)
	req.Header.Set("Authorization", "wrongauthorization")
	req.Header.Set("App", app)
	req.Header.Set("Operator", account)
	c.Request = req

	session, err := authenticator.Authenticate(c)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, constant.ErrUnauthorized, err)
}

func TestCalculateMD5(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "zentaozhangsanxxxx",
			expected: "7f18226c955beb483985aed5c581d0ba",
		},
		{
			input:    "test123",
			expected: calculateExpectedMD5("test123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := calculateMD5(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 辅助函数：计算期望的 MD5
func calculateExpectedMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
