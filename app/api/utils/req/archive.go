package req

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"github.com/gin-gonic/gin"
)

const (
	PathParamArchiveGitRef       = "*"
	QueryParamArchivePaths       = "path"
	QueryParamArchivePrefix      = "prefix"
	QueryParamArchiveAttributes  = "attributes"
	QueryParamArchiveTime        = "time"
	QueryParamArchiveCompression = "compression"
)

func Ext(path string) string {
	found := ""
	for _, format := range api.ArchiveFormats {
		if strings.HasSuffix(path, "."+string(format)) {
			if len(found) == 0 || len(found) < len(format) {
				found = string(format)
			}
		}
	}
	return found
}

func ParseArchiveParams(r *gin.Context) (api.ArchiveParams, string, error) {
	// 只支持 Path parameter 格式: /api/v2/repos/{id}/archive/{ref}.{format}
	//   Example: /api/v2/repos/32/archive/refs/heads/main.zip
	//            /api/v2/repos/32/archive/branch01.zip

	// 从 path 参数获取
	path := PathParamOrEmpty(r, "filePath")

	// If path is empty or only "/", try to extract from full request path
	if path == "" || path == "/" {
		// Fallback: parse from full URL path
		fullPath := r.Request.URL.Path
		archiveIdx := strings.Index(fullPath, "/archive/")
		if archiveIdx >= 0 {
			path = fullPath[archiveIdx+9:] // Skip "/archive/" (9 chars)
		}
	}

	// Remove leading slash that Gin adds to wildcard parameters
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		return api.ArchiveParams{}, "", errors.New("missing archive path, use format: /archive/{ref}.{format}")
	}

	// Split path into directory (rev) and filename
	rev, filename := filepath.Split(path)

	// Extract format from filename extension
	formatStr := Ext(filename)
	if formatStr == "" {
		return api.ArchiveParams{}, "", errors.Newf("cannot determine archive format from path '%s', supported formats: tar, zip, tar.gz, tgz", path)
	}

	// Build treeish by removing the format extension from filename
	name := strings.TrimSuffix(filename, "."+formatStr)
	treeish := rev + name

	// Parse and validate the format
	archFormat, err := api.ParseArchiveFormat(formatStr)
	if err != nil {
		return api.ArchiveParams{}, "", err
	}

	// Generate filename for response
	responseFilename := treeish + "." + formatStr
	// Clean filename to remove directory separators
	responseFilename = strings.ReplaceAll(responseFilename, "/", "_")

	// prefix is used for git archive to prefix all paths.
	prefix, _ := QueryParam(r, QueryParamArchivePrefix)
	attributes, _ := QueryParam(r, QueryParamArchiveAttributes)

	var mtime *time.Time
	timeStr, _ := QueryParam(r, QueryParamArchiveTime)
	if timeStr != "" {
		value, err := time.Parse(time.DateTime, timeStr)
		if err == nil {
			mtime = &value
		}
	}

	var compression *int
	compressionStr, _ := QueryParam(r, QueryParamArchiveCompression)
	if compressionStr != "" {
		value, err := strconv.Atoi(compressionStr)
		if err == nil {
			compression = &value
		}
	}

	return api.ArchiveParams{
		Format:      archFormat,
		Prefix:      prefix,
		Attributes:  api.ArchiveAttribute(attributes),
		Time:        mtime,
		Compression: compression,
		Treeish:     treeish,
		Paths:       r.Request.URL.Query()[QueryParamArchivePaths],
	}, responseFilename, nil
}
