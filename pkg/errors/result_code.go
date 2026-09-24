// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package errors

import (
	"net/http"

	"github.com/easysoft/gitfox-community/pkg/i18n"
)

const (
	NoType ErrorType = iota
	BadRequest
	NotFound
	DBError
	GitError
	SystemErr
	InternalError
	InvalidArgument
	ConfigNotFoundError
	Conflict
	PreconditionFailed
)

const defaultI18nKey = i18n.ErrorDefault

const defaultHttpStatus = http.StatusOK
