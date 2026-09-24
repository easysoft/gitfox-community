// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package common

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	dnsMatcher *regexp.Regexp = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*(?:xn--[a-zA-Z0-9\-]{1,59}|[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)$`)

	codeMatcher *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$`)
)

// IsIp check if the string is a ip address.
func IsIp(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	return ip != nil
}

// IsUrl check if the string is url.
func IsUrl(str string) bool {
	if str == "" {
		return false
	}

	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	allowedSchemes := map[string]struct{}{
		"http":   {},
		"https":  {},
		"ftp":    {},
		"ws":     {},
		"wss":    {},
		"file":   {},
		"mailto": {},
		"data":   {},
	}

	if _, ok := allowedSchemes[u.Scheme]; !ok {
		return false
	}

	if u.Scheme == "file" || u.Scheme == "mailto" || u.Scheme == "data" {
		return true
	}

	host := u.Hostname()
	if !strings.Contains(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}

	// domainRegexp := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$`)
	if !dnsMatcher.MatchString(host) {
		return false
	}

	return true
}

// IsDns check if the string is dns.
func IsDns(dns string) bool {
	return dnsMatcher.MatchString(dns)
}

// IsDbHost 判断是否是合法的数据库主机地址（域名或 IP，不含端口和协议）
func IsDbHost(host string) bool {
	return IsIp(host) || IsDns(host)
}

// IsValidCode 校验标识符是否符合规范：首尾字符必须是字母或数字，中间可包含字母、数字、点、下划线、中划线。
func IsValidCode(code string) bool {
	return codeMatcher.MatchString(code)
}
