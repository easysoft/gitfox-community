// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package auth

import (
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
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
