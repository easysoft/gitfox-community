package authn

import (
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox/types/enum"
	"github.com/golang-jwt/jwt"
)

const (
	issuer = "Gitfox"
)

// Source represents the source of the SubClaimsAccessPermissions.
type Source string

const (
	OciSource Source = "oci"
)

// Claims defines Harness jwt claims.
type Claims struct {
	jwt.StandardClaims

	App        string `json:"app,omitempty"`
	Admin      bool   `json:"admin,omitempty"`
	Account    string `json:"account,omitempty"`
	Identifier string `json:"identifier,omitempty"`

	Token             *SubClaimsToken             `json:"tkn,omitempty"`
	Membership        *SubClaimsMembership        `json:"ms,omitempty"`
	AccessPermissions *SubClaimsAccessPermissions `json:"ap,omitempty"`
}

// SubClaimsToken contains information about the token the JWT was created for.
type SubClaimsToken struct {
	Type constant.TokenType `json:"typ,omitempty"`
	ID   int64              `json:"id,omitempty"`
}

// SubClaimsMembership contains the ephemeral membership the JWT was created with.
type SubClaimsMembership struct {
	Role    enum.MembershipRole `json:"role,omitempty"`
	SpaceID int64               `json:"sid,omitempty"`
}

// SubClaimsAccessPermissions stores allowed actions on a resource.
type SubClaimsAccessPermissions struct {
	Source      Source              `json:"src,omitempty"`
	Permissions []AccessPermissions `json:"permissions,omitempty"`
}

// AccessPermissions stores allowed actions on a resource.
type AccessPermissions struct {
	SpaceID     int64             `json:"sid,omitempty"`
	Permissions []enum.Permission `json:"p"`
}

// GenerateForToken generates a jwt for a given token.
func GenerateForToken(token *auth.Token, secret string) (string, error) {
	var expiresDate int64
	if token.ExpirationDate != nil {
		expiresDate = token.ExpirationDate.Unix()
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec not millisec
			IssuedAt:  token.CreatedDate.Unix(),
			ExpiresAt: expiresDate / 1000,
		},
		Account:    token.Account,
		Identifier: token.Identifier,
		Token: &SubClaimsToken{
			Type: token.Type,
			ID:   token.ID,
		},
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return res, nil
}

// GenerateForToken generates a jwt for a given token.
func GenerateForAppToken(token *auth.Token, secret string, isAdmin bool) (string, error) {
	var expiresDate int64
	if token.ExpirationDate != nil {
		expiresDate = token.ExpirationDate.Unix()
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec not millisec
			IssuedAt:  token.CreatedDate.Unix(),
			ExpiresAt: expiresDate / 1000,
		},
		App:        token.Account,
		Admin:      isAdmin,
		Account:    token.Account,
		Identifier: token.Identifier,

		Token: &SubClaimsToken{
			Type: token.Type,
			ID:   token.ID,
		},
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign app token: %w", err)
	}

	return res, nil
}

// GenerateWithMembership generates a jwt with the given ephemeral membership.
func GenerateWithMembership(account string, spaceID int64, role enum.MembershipRole, lifetime time.Duration, secret string) (string, error) {
	issuedAt := time.Now()
	expiresDate := issuedAt.Add(lifetime)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expiresDate.Unix(),
		},
		Account: account,
		Membership: &SubClaimsMembership{
			SpaceID: spaceID,
			Role:    role,
		},
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return res, nil
}

// GenerateForTokenWithAccessPermissions generates a jwt for a given token.
func GenerateForTokenWithAccessPermissions(account string, lifetime *time.Duration, secret string, accessPermissions *SubClaimsAccessPermissions) (string, error) {
	issuedAt := time.Now()
	if lifetime == nil {
		return "", fmt.Errorf("token lifetime is required")
	}
	expiresDate := issuedAt.Add(*lifetime)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    issuer,
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expiresDate.Unix(),
		},
		Account:           account,
		AccessPermissions: accessPermissions,
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return res, nil
}

func KeysToSecret(key ...string) string {
	return strings.Join(key, "_")
}
