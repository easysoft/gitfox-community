package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims JWT声明
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// 签名密钥(实际项目中应从配置读取)
var secretKey = []byte("your-secret-key")

// GenerateToken 生成JWT令牌
func GenerateToken(userID string) (string, error) {
	// 设置过期时间(24小时)
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "devops-platform",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	return token.SignedString(secretKey)
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	// 验证令牌并返回声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
