package authn

import (
	"fmt"
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

	PrincipalID int64 `json:"pid,omitempty"`

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
	var expiresAt int64
	if token.ExpiresAt != nil {
		expiresAt = token.ExpiresAt.Unix()
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec not millisec
			IssuedAt:  token.CreatedDate.Unix(),
			ExpiresAt: expiresAt / 1000,
		},
		PrincipalID: token.PrincipalID,
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

// GenerateWithMembership generates a jwt with the given ephemeral membership.
func GenerateWithMembership(
	principalID int64,
	spaceID int64,
	role enum.MembershipRole,
	lifetime time.Duration,
	secret string,
) (string, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(lifetime)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expiresAt.Unix(),
		},
		PrincipalID: principalID,
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
func GenerateForTokenWithAccessPermissions(
	principalID int64,
	lifetime *time.Duration,
	secret string, accessPermissions *SubClaimsAccessPermissions,
) (string, error) {
	issuedAt := time.Now()
	if lifetime == nil {
		return "", fmt.Errorf("token lifetime is required")
	}
	expiresAt := issuedAt.Add(*lifetime)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    issuer,
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expiresAt.Unix(),
		},
		PrincipalID:       principalID,
		AccessPermissions: accessPermissions,
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return res, nil
}
