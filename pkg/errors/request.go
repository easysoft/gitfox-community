// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package errors

// Request related errors
var (
	// ErrParameterMissing 表示缺少必需的参数
	ErrParameterMissing = WithI18nKey(BadRequest.New("parameter missing"), "request.parameter_missing")

	// ErrParameterInvalid 表示参数无效
	ErrParameterInvalid = WithI18nKey(BadRequest.New("parameter invalid"), "request.parameter_invalid")

	// ErrPathParameterNotFound 表示路径参数未找到
	ErrPathParameterNotFound = WithI18nKey(BadRequest.New("path parameter not found"), "request.path_parameter_not_found")

	// ErrQueryParameterNotFound 表示查询参数未找到
	ErrQueryParameterNotFound = WithI18nKey(BadRequest.New("query parameter not found"), "request.query_parameter_not_found")

	// ErrParameterDecodeFailed 表示参数解码失败
	ErrParameterDecodeFailed = WithI18nKey(BadRequest.New("parameter decode failed"), "request.parameter_decode_failed")

	// ErrParameterTypeMismatch 表示参数类型不匹配
	ErrParameterTypeMismatch = WithI18nKey(BadRequest.New("parameter type mismatch"), "request.parameter_type_mismatch")

	// ErrInvalidIntegerParameter 表示整数参数无效
	ErrInvalidIntegerParameter = WithI18nKey(BadRequest.New("parameter must be a valid integer"), "request.invalid_integer_parameter")

	// ErrInvalidPositiveIntegerParameter 表示正整数参数无效
	ErrInvalidPositiveIntegerParameter = WithI18nKey(BadRequest.New("parameter must be a positive integer"), "request.invalid_positive_integer_parameter")

	// ErrInvalidBooleanParameter 表示布尔参数无效
	ErrInvalidBooleanParameter = WithI18nKey(BadRequest.New("parameter must be a boolean"), "request.invalid_boolean_parameter")

	// ErrInvalidRequestBody 表示请求体无效
	ErrInvalidRequestBody = WithI18nKey(BadRequest.New("invalid request body"), "request.invalid_request_body")

	// ErrInvalidJSON 表示JSON格式无效
	ErrInvalidJSON = WithI18nKey(BadRequest.New("invalid JSON format"), "request.invalid_json_format")

	// ErrInvalidContentType 表示Content-Type无效
	ErrInvalidContentType = WithI18nKey(BadRequest.New("invalid content type"), "request.invalid_content_type")

	// ErrRequestTooLarge 表示请求体过大
	ErrRequestTooLarge = WithI18nKey(BadRequest.New("request body too large"), "request.request_body_too_large")

	// ErrInvalidURLParameter 表示URL参数无效
	ErrInvalidURLParameter = WithI18nKey(BadRequest.New("invalid URL parameter"), "request.invalid_url_parameter")
)
