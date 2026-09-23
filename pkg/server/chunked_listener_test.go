// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package server

import (
	"bytes"
	"strings"
	"testing"
)

// TestPatchHeader_DockerChunkedPatch 复现 docker 28 client PATCH 漏带 TE 头的真实场景，
// 期望 patchHeader 注入 "Transfer-Encoding: chunked"。
func TestPatchHeader_DockerChunkedPatch(t *testing.T) {
	head := buildHead(
		"PATCH /v2/global/global/backend/pipeline-helper/blobs/uploads/abc?_state=xx HTTP/1.1",
		"Host: 10.0.4.2:3000",
		"User-Agent: docker/28.4.0",
		"Authorization: Basic YWRtaW46Y2hhbmdlaXQ=",
		"Content-Type: application/octet-stream",
		"Accept-Encoding: gzip",
	)

	got := patchHeader(head)
	if !bytes.Contains(got, []byte("\r\nTransfer-Encoding: chunked\r\n")) {
		t.Fatalf("expected Transfer-Encoding to be injected, got:\n%s", got)
	}
	if !bytes.HasSuffix(got, []byte("\r\n\r\n")) {
		t.Fatalf("expected header to end with CRLFCRLF, got:\n%s", got)
	}
}

// TestPatchHeader_KeepCLUntouched: 已经带 Content-Length 的 PUT 不应被改写。
func TestPatchHeader_KeepCLUntouched(t *testing.T) {
	head := buildHead(
		"PUT /v2/global/foo/bar/blobs/uploads/abc?digest=sha256:xx HTTP/1.1",
		"Host: 10.0.4.2:3000",
		"Content-Length: 0",
	)
	got := patchHeader(head)
	if !bytes.Equal(got, head) {
		t.Fatalf("expected no change, diff:\nwant=%s\ngot =%s", head, got)
	}
}

// TestPatchHeader_KeepTEUntouched: 已带 Transfer-Encoding 的 PATCH 不应被重复注入。
func TestPatchHeader_KeepTEUntouched(t *testing.T) {
	head := buildHead(
		"PATCH /v2/global/foo/bar/blobs/uploads/abc HTTP/1.1",
		"Transfer-Encoding: chunked",
		"Host: 10.0.4.2:3000",
	)
	got := patchHeader(head)
	if !bytes.Equal(got, head) {
		t.Fatalf("expected no change, diff:\nwant=%s\ngot =%s", head, got)
	}
}

// TestPatchHeader_NotMatchingPath: 不在目标路径上的请求不应被改写。
func TestPatchHeader_NotMatchingPath(t *testing.T) {
	head := buildHead(
		"PATCH /api/v2/users/1 HTTP/1.1",
		"Host: 10.0.4.2:3000",
		"Content-Type: application/json",
	)
	got := patchHeader(head)
	if !bytes.Equal(got, head) {
		t.Fatalf("expected no change for non-OCI path, got:\n%s", got)
	}
}

// TestPatchHeader_GETIgnored: GET 即便命中路径也不应被改写。
func TestPatchHeader_GETIgnored(t *testing.T) {
	head := buildHead(
		"GET /v2/global/foo/bar/blobs/uploads/abc HTTP/1.1",
		"Host: 10.0.4.2:3000",
	)
	got := patchHeader(head)
	if !bytes.Equal(got, head) {
		t.Fatalf("expected GET unchanged, got:\n%s", got)
	}
}

// TestParseBodyLen_AfterPatch: 注入 TE 后 parseBodyLen 应返回 -1。
func TestParseBodyLen_AfterPatch(t *testing.T) {
	head := buildHead(
		"PATCH /v2/global/foo/bar/blobs/uploads/abc HTTP/1.1",
		"Host: 10.0.4.2:3000",
	)
	got := patchHeader(head)
	if n := parseBodyLen(got); n != -1 {
		t.Fatalf("after patch parseBodyLen want -1, got %d", n)
	}
}

// TestParseBodyLen_CL: Content-Length 头应被正确解析为数值。
func TestParseBodyLen_CL(t *testing.T) {
	head := buildHead(
		"POST /api/v2/foo HTTP/1.1",
		"Content-Length: 1234",
	)
	if n := parseBodyLen(head); n != 1234 {
		t.Fatalf("want 1234, got %d", n)
	}
}

// TestParseBodyLen_NoBody: 既无 CL 又无 TE 应返回 0。
func TestParseBodyLen_NoBody(t *testing.T) {
	head := buildHead(
		"GET /v2/ HTTP/1.1",
		"Host: x",
	)
	if n := parseBodyLen(head); n != 0 {
		t.Fatalf("want 0, got %d", n)
	}
}

func buildHead(lines ...string) []byte {
	return []byte(strings.Join(lines, "\r\n") + "\r\n\r\n")
}
