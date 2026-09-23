// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// MaxResponseBodySize 最大响应体大小限制 (16MB)
	MaxResponseBodySize int64 = 16 * 1024 * 1024
)

var httpcli = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   20 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     2 * time.Minute,
	},
	Timeout: 20 * time.Second,
}

// BaseResult 基础返回体
type BaseResult struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Get 发送 GET 请求
func Get(ctx context.Context, url string, headers, cookies map[string]string, resp interface{}) (interface{}, error) {
	result := BaseResult{}
	if err := Request(ctx, "GET", url, headers, cookies, nil, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// Post 发送 POST 请求
func Post(ctx context.Context, url string, headers, cookies map[string]string, data interface{}, resp interface{}) (interface{}, error) {
	result := BaseResult{}
	if err := Request(ctx, "POST", url, headers, cookies, data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// Request 执行 HTTP 请求
func Request(ctx context.Context, method string, url string, headers, cookies map[string]string, data interface{}, result interface{}) error {
	var req *http.Request
	var err error

	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH":
		var reqBody []byte
		if data != nil {
			reqBody, err = json.Marshal(data)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
		}
		req, err = http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create %s request: %w", method, err)
		}
		// 只为有请求体的方法设置 Content-Type
		req.Header.Set("Content-Type", "application/json")
	case "GET", "DELETE":
		req, err = http.NewRequestWithContext(ctx, strings.ToUpper(method), url, nil)
		if err != nil {
			return fmt.Errorf("failed to create %s request: %w", method, err)
		}
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// 添加 cookie
	for k, v := range cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	// 添加通用请求头
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	response, err := httpcli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		// 确保响应体被完全读取并关闭
		if response.Body != nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
	}()

	// 检查 HTTP 状态码
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	// 如果不需要解析结果，直接返回
	if result == nil {
		return nil
	}

	// 安全地读取响应体
	// 使用 LimitReader 防止读取过大的响应体
	limitedReader := io.LimitReader(response.Body, MaxResponseBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查是否超过大小限制
	if int64(len(body)) >= MaxResponseBodySize {
		return fmt.Errorf("response body exceeds maximum size of %d bytes", MaxResponseBodySize)
	}

	// 解析 JSON 响应
	if err = json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
