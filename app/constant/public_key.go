// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

// PublicKeyUsage represents usage type of public key.
type PublicKeyUsage string

// PublicKeyUsage enumeration.
const (
	PublicKeyUsageAuth PublicKeyUsage = "auth"
	PublicKeyUsageSign PublicKeyUsage = "sign"
)
