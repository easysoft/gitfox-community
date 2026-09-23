// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package authn

import "fmt"

type authKind string

const (
	AuthKindEasy  = "easy"
	AuthKindBasic = "basic"
)

var _authMethods = make(map[authKind]Authenticator)

func Register(kind authKind, authenticator Authenticator) error {
	if _, ok := _authMethods[kind]; ok {
		return fmt.Errorf("conflict authenticator '%s'", kind)
	}
	_authMethods[kind] = authenticator
	return nil
}

func Get(kind authKind) Authenticator {
	return _authMethods[kind]
}
