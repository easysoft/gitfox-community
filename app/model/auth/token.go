package auth

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type Token struct {
	ID             int64              `json:"id"`
	Account        string             `json:"account"`
	Type           constant.TokenType `json:"type"`
	Identifier     string             `json:"identifier"`
	ExpirationDate *time.Time         `json:"expirationDate"`

	CreatedDate *time.Time `json:"created_date"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Token       Token  `json:"token"`
}

type LoginResp struct {
	Token *TokenResponse `json:"token"`
	PAT   *TokenResponse `json:"pat"`
}
