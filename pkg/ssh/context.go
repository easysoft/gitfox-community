// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package ssh

import (
	"context"

	"github.com/gliderlabs/ssh"
)

type ctxSessionKey struct{}

func WithContext(ctx context.Context, session ssh.Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey{}, session)
}

func FromContext(ctx context.Context) ssh.Session {
	obj, ok := ctx.Value(ctxSessionKey{}).(ssh.Session)
	if !ok {
		return nil
	}
	return obj
}
