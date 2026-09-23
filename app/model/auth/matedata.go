// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package auth

import "github.com/easysoft/gitfox-community/app/constant"

type Metadata interface {
	ImpactsAuthorization() bool
}

// EmptyMetadata represents the state when the auth session doesn't have any extra metadata.
type EmptyMetadata struct{}

func (m *EmptyMetadata) ImpactsAuthorization() bool {
	return false
}

// TokenMetadata contains information about the token that was used during auth.
type TokenMetadata struct {
	TokenType constant.TokenType
	TokenID   int64
}

func (m *TokenMetadata) ImpactsAuthorization() bool {
	return false
}
