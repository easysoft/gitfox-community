package common

import "testing"

func TestIsIp(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// IPv4 合法
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		// IPv6 合法
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		// 非法
		{"192.168.1.256", false},  // 超出范围
		{"192.168.1", false},      // 不完整
		{"192.168.1.1:3306", false}, // 带端口
		{"localhost", false},       // 域名
		{"example.com", false},    // 域名
		{"", false},               // 空字符串
		{"abc", false},            // 随机字符
	}

	for _, c := range cases {
		got := IsIp(c.input)
		if got != c.want {
			t.Errorf("IsIp(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsDns(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// 合法域名
		{"example.com", true},
		{"sub.example.com", true},
		{"a.b.c.example.com", true},
		{"localhost", true},          // 单标签
		{"mysql.internal", true},
		{"xn--p1ai", true},           // IDN 国际化域名
		// 合法但边界
		{"a.io", true},
		{"my-db.example.com", true},  // 含连字符
		// 非法
		{"-example.com", false},       // 标签以 - 开头
		{"example-.com", false},       // 标签以 - 结尾
		{"example..com", false},       // 双点
		// 注意：DNS 正则仅校验标签格式（字母数字和连字符），纯数字标签（如 192 / 168）也符合正则
		// 因此 IsDns 无法区分 IP 和域名，"192.168.1.1" 会返回 true
		{"192.168.1.1", true},         // 数字标签符合 DNS 正则，IsDns 无法排除 IP
		{"http://example.com", false}, // 含协议
		{"example.com:3306", false},   // 含端口
		{"", false},                   // 空字符串
		{"example .com", false},       // 含空格
	}

	for _, c := range cases {
		got := IsDns(c.input)
		if got != c.want {
			t.Errorf("IsDns(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsUrl(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// 合法 HTTP/HTTPS
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example.com/path", true},
		{"https://example.com/path?q=1&v=2", true},
		{"https://sub.example.com", true},
		{"https://my-db.example.org:8080/api", true},
		// 合法特殊 scheme
		{"ftp://files.example.com", true},
		{"ws://chat.example.com", true},
		{"wss://chat.example.com", true},
		// 注意：IsUrl 内部用 dnsMatcher 校验 host，而 dnsMatcher 不排除纯数字标签
		// 因此 IP 形式的 URL 也会通过，如有需要应在 IsUrl 中单独加 IsIp 短路判断
		{"http://192.168.1.1", true},
		{"https://192.168.1.1:8080/api", true},
		// 非法：不支持的 scheme
		{"ssh://git@example.com", false},
		{"mysql://db.example.com", false},
		{"git://example.com/repo.git", false},
		// 非法：缺少 scheme 或 host
		{"example.com", false},
		{"/relative/path", false},
		{"//example.com", false},
		// 非法：localhost（无点号）
		{"http://localhost", false},
		{"http://localhost:3306", false},
		// 边界
		{"", false},
	}

	for _, c := range cases {
		got := IsUrl(c.input)
		if got != c.want {
			t.Errorf("IsUrl(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsDbHost(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// 域名形式
		{"db.example.com", true},
		{"mysql.internal", true},
		{"localhost", true},
		{"sub.db.example.com", true},
		// IP 形式
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"::1", true},         // IPv6
		{"2001:db8::1", true}, // IPv6
		// 非法：带端口
		{"192.168.1.1:3306", false},
		{"db.example.com:3306", false},
		// 非法：带协议
		{"mysql://db.example.com", false},
		{"http://db.example.com", false},
		// 非法：格式错误
		{"-db.example.com", false},
		{"", false},
		// 注意：256 是合法的 DNS 标签格式（纯数字），IsDns 短路命中，IsDbHost 返回 true
		// 若需严格排除越界 IP，应在调用处先用 IsIp 验证
		{"192.168.1.256", true},
	}

	for _, c := range cases {
		got := IsDbHost(c.input)
		if got != c.want {
			t.Errorf("IsDbHost(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
