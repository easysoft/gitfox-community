// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

import (
	"net/http"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

// atuh
var (
	ErrAuthNoData = errors.WithI18nKey(errors.DBError.New("the request doesn't contain any auth data that can be used by the Authorizer"), i18n.AuthNoData)

	// ErrUnauthorized is returned when the acting principal is not authenticated.
	ErrUnauthorized = errors.WithHttpStatus(errors.InternalError.New("the acting principal is not authenticated"), http.StatusUnauthorized)

	// ErrUserPass is returned when a resource is not found.
	ErrUserPass = errors.WithHttpStatus(errors.InternalError.New("Invalid username or password"), http.StatusUnauthorized)
)

// system
var (
	// ErrInternal is returned when an internal error occurred.
	ErrInternal = errors.WithHttpStatus(errors.InternalError.New("Internal error occurred"), http.StatusInternalServerError)

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.WithHttpStatus(errors.NotFound.New("Not Found"), http.StatusNotFound)

	// ErrBadRequest is returned when there was an issue with the input.
	ErrBadRequest = errors.WithHttpStatus(errors.BadRequest.New("Bad Request"), http.StatusBadRequest)
)
