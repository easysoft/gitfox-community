// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package req

import (
	"context"
	"net/url"
	"reflect"
	"strconv"

	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/pkg/errors"

	"github.com/gin-gonic/gin"
)

// SetAuthSession returns a copy of parent in which the principal
// value is set.
func SetAuthSession(parent context.Context, v *auth.Session) context.Context {
	return context.WithValue(parent, authSessionKey, v)
}

// GetAuthSession returns the value of the principal key on the
// context.
func GetAuthSession(ctx context.Context) (*auth.Session, bool) {
	v, ok := ctx.Value(authSessionKey).(*auth.Session)
	return v, ok && v != nil
}

func GetOperator(ctx context.Context) string {
	var operator string
	if session, ok := GetAuthSession(ctx); ok && session != nil && session.User != nil {
		operator = session.User.Account
	}
	return operator
}

func GetAuthSessionWithEmail(ctx context.Context) (*auth.Session, bool) {
	v, ok := ctx.Value(authSessionKey).(*auth.Session)
	if ok && v != nil && v.User.Email != "" {
		return v, true
	}
	return v, false
}

// GetCookie tries to retrieve the cookie from the request or returns false if it doesn't exist.
func GetCookie(c *gin.Context, cookieName string) (string, bool) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return "", false
	}

	return cookie, true
}

// GetHeaderOrDefault returns the value of the first non-empty header occurrence.
// If no value is found, the default value is returned.
func GetHeaderOrDefault(c *gin.Context, headerName string, dflt string) string {
	val, ok := GetHeader(c, headerName)
	if !ok {
		return dflt
	}

	return val
}

// GetHeader returns the value of the first non-empty header occurrence.
// If no value is found, `false` is returned.
func GetHeader(c *gin.Context, headerName string) (string, bool) {
	val := c.GetHeader(headerName)
	if val == "" {
		return "", false
	}

	return val, true
}

// PathParamOrError tries to retrieve the parameter from the request and
// returns the parameter if it exists and is not empty, otherwise returns an error.
func PathParamOrError(c *gin.Context, paramName string) (string, error) {
	val, err := PathParam(c, paramName)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", errors.BadRequest.Newf("Parameter '%s' not found in request path.", paramName)
	}

	return val, nil
}

// PathParamOrEmpty retrieves the path parameter or returns an empty string otherwise.
func PathParamOrEmpty(c *gin.Context, paramName string) string {
	val, err := PathParam(c, paramName)
	if err != nil {
		return ""
	}

	return val
}

// PathParam retrieves the path parameter or returns false if it exists.
func PathParam(c *gin.Context, paramName string) (string, error) {
	val := c.Param(paramName)
	if val == "" {
		return "", nil
	}

	val, err := url.PathUnescape(val)
	if err != nil {
		return "", errors.BadRequest.Newf("Failed to decode path parameter '%s'.", paramName)
	}

	return val, nil
}

// QueryParam returns the parameter if it exists.
func QueryParam(c *gin.Context, paramName string) (string, bool) {
	val, exists := c.GetQuery(paramName)
	return val, exists
}

// QueryParamList returns list of the parameter values if they exist.
func QueryParamList(c *gin.Context, paramName string) ([]string, bool) {
	query := c.Request.URL.Query()
	if !query.Has(paramName) {
		return nil, false
	}

	return query[paramName], true
}

// QueryParamOrDefault retrieves the parameter from the query and
// returns the parameter if it exists, otherwise returns the provided default value.
func QueryParamOrDefault(c *gin.Context, paramName string, deflt string) string {
	val, ok := QueryParam(c, paramName)
	if !ok {
		return deflt
	}

	return val
}

// QueryParamOrError tries to retrieve the parameter from the query and
// returns the parameter if it exists, otherwise returns an error.
func QueryParamOrError(c *gin.Context, paramName string) (string, error) {
	val, ok := QueryParam(c, paramName)
	if !ok {
		return "", errors.BadRequest.Newf("Parameter '%s' not found in query.", paramName)
	}

	return val, nil
}

// QueryParamAsPositiveInt64 extracts an integer parameter from the request query.
// If the parameter doesn't exist the provided default value is returned.
func QueryParamAsPositiveInt64OrDefault(c *gin.Context, paramName string, deflt int64) (int64, error) {
	value, ok := QueryParam(c, paramName)
	if !ok {
		return deflt, nil
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil || valueInt <= 0 {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return valueInt, nil
}

// QueryParamAsPositiveInt64OrError extracts an integer parameter from the request query.
// If the parameter doesn't exist an error is returned.
func QueryParamAsPositiveInt64OrError(c *gin.Context, paramName string) (int64, error) {
	value, err := QueryParamOrError(c, paramName)
	if err != nil {
		return 0, err
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil || valueInt <= 0 {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return valueInt, nil
}

// QueryParamAsPositiveInt64 extracts an integer parameter from the request query if it exists.
func QueryParamAsPositiveInt64(c *gin.Context, paramName string) (int64, bool, error) {
	value, ok := QueryParam(c, paramName)
	if !ok {
		return 0, false, nil
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil || valueInt <= 0 {
		return 0, false, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return valueInt, true, nil
}

// PathParamAsPositiveInt64 extracts an integer parameter from the request path.
func PathParamAsPositiveInt64(c *gin.Context, paramName string) (int64, error) {
	rawValue, err := PathParamOrError(c, paramName)
	if err != nil {
		return 0, err
	}

	valueInt, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil || valueInt <= 0 {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return valueInt, nil
}

// PathParamAsPositiveUint16 extracts an integer parameter from the request path.
func PathParamAsPositiveUint16(c *gin.Context, paramName string) (uint16, error) {
	rawValue, err := PathParamOrError(c, paramName)
	if err != nil {
		return 0, err
	}

	valueInt, err := strconv.ParseInt(rawValue, 10, 0)
	if err != nil || valueInt <= 0 {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return uint16(valueInt), nil
}

// PathParamAsPositiveUint extracts an integer parameter from the request path.
func PathParamAsPositiveUint(c *gin.Context, paramName string) (uint, error) {
	rawValue, err := PathParamOrError(c, paramName)
	if err != nil {
		return 0, err
	}

	valueUint, err := strconv.ParseUint(rawValue, 10, 0)
	if err != nil || valueUint <= 0 {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a positive integer.", paramName)
	}

	return uint(valueUint), nil
}

// PathParamAsNaturalNum extracts an integer parameter from the request path.
func PathParamAsNaturalNum(c *gin.Context, paramName string) (uint, error) {
	rawValue, err := PathParamOrError(c, paramName)
	if err != nil {
		return 0, err
	}

	valueUint, err := strconv.ParseUint(rawValue, 10, 0)
	if err != nil {
		return 0, errors.BadRequest.Newf("Parameter '%s' must be a integer.", paramName)
	}

	return uint(valueUint), nil
}

// QueryParamAsBoolOrDefault tries to retrieve the parameter from the query and parse it to bool.
func QueryParamAsBoolOrDefault(c *gin.Context, paramName string, deflt bool) (bool, error) {
	rawValue, ok := QueryParam(c, paramName)
	if !ok || len(rawValue) == 0 {
		return deflt, nil
	}

	boolValue, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, errors.BadRequest.Newf("Parameter '%s' must be a boolean.", paramName)
	}

	return boolValue, nil
}

// QueryParamListAsPositiveInt64 extracts integer parameter slice from the request query.
func QueryParamListAsPositiveInt64(c *gin.Context, paramName string) ([]int64, error) {
	valuesString, ok := QueryParamList(c, paramName)
	if !ok {
		return make([]int64, 0), nil
	}
	valuesInt := make([]int64, len(valuesString))

	for i, vs := range valuesString {
		vi, err := strconv.ParseInt(vs, 10, 64)
		if err != nil || vi <= 0 {
			return nil, errors.BadRequest.Newf("Parameter %q must be a positive integer.", paramName)
		}
		valuesInt[i] = vi
	}
	return valuesInt, nil
}

// SetDefaults sets default values for struct fields using the `default` tag.
// The function uses reflection to iterate through struct fields and set defaults
// for zero-valued fields that have a `default` tag.
// It also recursively processes embedded (anonymous) structs.
//
// Supported field types: int, uint, string, bool, float
//
// Example usage:
//
//	type ReqFilter struct {
//	    Page  int    `json:"page" default:"1"`
//	    Size  int    `json:"size" default:"20"`
//	    Sort  string `json:"sort" default:"name"`
//	    Order string `json:"order" default:"asc"`
//	}
//
//	var req ReqFilter
//	c.ShouldBindJSON(&req)
//	SetDefaults(&req)
func SetDefaults(v interface{}) {
	val := reflect.ValueOf(v)

	// Check if the input is a pointer
	if val.Kind() != reflect.Ptr {
		return
	}

	// Get the element the pointer points to
	val = val.Elem()

	// Check if it's a struct
	if val.Kind() != reflect.Struct {
		return
	}

	setStructDefaults(val)
}

// setStructDefaults is a helper function that processes a struct value
func setStructDefaults(val reflect.Value) {
	typ := val.Type()

	// Iterate through all fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// If the field is an embedded (anonymous) struct, recursively process it
		if typeField.Anonymous && field.Kind() == reflect.Struct {
			setStructDefaults(field)
			continue
		}

		// Get the default tag value
		defaultVal := typeField.Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		// Only set default if the field is zero-valued
		if !field.IsZero() {
			continue
		}

		// Set default value based on field type
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
				field.SetInt(intVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintVal, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
				field.SetUint(uintVal)
			}
		case reflect.String:
			field.SetString(defaultVal)
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(defaultVal); err == nil {
				field.SetBool(boolVal)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(defaultVal, 64); err == nil {
				field.SetFloat(floatVal)
			}
		}
	}
}
