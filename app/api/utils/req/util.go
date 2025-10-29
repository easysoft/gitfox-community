package req

import (
	"net/url"
	"strconv"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/gin-gonic/gin"
)

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
