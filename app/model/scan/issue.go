package scan

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// ScanIssue 扫描问题内部模型
type ScanIssue struct {
	Key             string                   `json:"key"`
	RuleID          int64                    `json:"ruleID"`
	Message         string                   `json:"message"`
	Path            string                   `json:"path"`
	Line            int64                    `json:"line"`
	RepoID          int64                    `json:"repoID"`
	RepoBranch      string                   `json:"repoBranch"`
	CreatedByTaskID int64                    `json:"createdByTaskID"`
	UpdatedByTaskID int64                    `json:"updatedByTaskID"`
	Status          constant.ScanIssueStatus `json:"status"`
	ScanMethod      constant.ScanMethod      `json:"scanMethod"`
	Payload         json.RawMessage          `json:"payload"`
}

// SetPayload 设置扩展数据
func (i *ScanIssue) SetPayload(payload interface{}) error {
	if payload == nil {
		i.Payload = nil
		return nil
	}
	var err error
	if i.Payload, err = json.Marshal(payload); err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return nil
}

// ScanIssueKey 已有问题的 key 映射
type ScanIssueKey struct {
	ID              int64                    `json:"id"`
	Key             string                   `json:"key"`
	CreatedByTaskID int64                    `json:"createdByTaskID"`
	UpdatedByTaskID int64                    `json:"updatedByTaskID"`
	Status          constant.ScanIssueStatus `json:"status"`
}

// ScanIssueReport 扫描报告统计
type ScanIssueReport struct {
	Total       int `json:"total"`
	Add         int `json:"add"`
	Update      int `json:"update"`
	Resolve     int `json:"resolve"`
	Conflict    int `json:"conflict"`
	Skip        int `json:"skip"`
	MissingRule int `json:"missingRule"`
}

// ScanIssueComparedData issue 比对结果
type ScanIssueComparedData struct {
	Added   []*ScanIssue
	Updated []int64
	Removed []int64
	All     []*ScanIssue
}

// IssueLocation 问题位置信息
type IssueLocation struct {
	Path  string `json:"path"`
	Range struct {
		StartLine int64 `json:"startLine"`
		EndLine   int64 `json:"endLine"`
		StartByte int64 `json:"startByte,omitempty"`
		EndByte   int64 `json:"endByte,omitempty"`
	} `json:"range"`
	Commit struct {
		CommitID       string `json:"commit_id"`
		CommitShortID  string `json:"commit_short_id"`
		Summary        string `json:"summary"`
		Author         string `json:"author"`
		AuthorEmail    string `json:"author_email"`
		AuthorDate     int64  `json:"author_date"`
		Committer      string `json:"committer"`
		CommitterEmail string `json:"committer_email"`
		CommitDate     int64  `json:"committer_date"`
	} `json:"commit"`
}

// IssueInputBase 问题输入基础结构
type IssueInputBase struct {
	RuleKey  string `json:"ruleKey"`
	Message  string `json:"message"`
	Level    string `json:"level"`
	Language string `json:"language"`
}

// IssuePayloadCheck check 类型扩展数据
type IssuePayloadCheck struct {
	Tool               string        `json:"tool"`
	Category           string        `json:"category"`
	Snippet            string        `json:"snippet"`
	SnippetWithContext string        `json:"snippetWithContext"`
	Location           IssueLocation `json:"location"`
	Mode               string        `json:"mode"`
}

// IssuePayloadSmell smell 类型扩展数据
type IssuePayloadSmell struct {
	Driver             string          `json:"driver,omitempty"`
	Tool               string          `json:"tool"`
	Category           string          `json:"category"`
	Snippet            string          `json:"snippet"`
	SnippetWithContext string          `json:"snippetWithContext"`
	Location           IssueLocation   `json:"location"`
	OtherLocations     []IssueLocation `json:"otherLocations,omitempty"`
	Mode               string          `json:"mode"`
}

// IssueInputCheck check 类型输入
type IssueInputCheck struct {
	IssueInputBase
	IssuePayloadCheck
}

// UniqueKey 生成唯一标识
func (t IssueInputCheck) UniqueKey() string {
	return fmt.Sprintf("%s/%s/%d/%d,'%s',`%s`", t.RuleKey, t.Location.Path, t.Location.Range.StartLine, t.Location.Range.EndLine, t.Message, t.Snippet)
}

// IssueInputSmell smell 类型输入
type IssueInputSmell struct {
	IssueInputBase
	IssuePayloadSmell
}

// UniqueKey 生成唯一标识
func (t IssueInputSmell) UniqueKey() string {
	return fmt.Sprintf("%s/%s/%d/%d,'%s',`%s`", t.RuleKey, t.Location.Path, t.Location.Range.StartLine, t.Location.Range.EndLine, t.Message, t.Snippet)
}

// GenerateIssueKey 根据任务信息生成 issue 唯一 key
func GenerateIssueKey(repoID int64, executionTarget, part string) string {
	in := []byte(fmt.Sprintf("%s,%d,%s", part, repoID, executionTarget))
	h := sha1.New()
	h.Write(in)
	return hex.EncodeToString(h.Sum(nil))
}

// RulesMap 规则 key 到 rule ID 的映射
type RulesMap map[string]int64

// GetByTool 根据工具名和 ruleKey 查找规则 ID
func (m RulesMap) GetByTool(tool, ruleKey string) (int64, bool) {
	lang := mapToolToLang(tool)
	if lang == "" {
		return 0, false
	}
	indexKey := fmt.Sprintf("%s:%s", lang, ruleKey)
	value, ok := m[indexKey]
	return value, ok
}

// mapToolToLang 将工具名映射为语言标识
func mapToolToLang(tool string) string {
	switch tool {
	case "phpstan", "php-codesniffer":
		return "php"
	case "bandit", "black", "flake8", "ruff":
		return "python"
	case "brakeman", "rubocop", "standardrb":
		return "ruby"
	case "clippy":
		return "rust"
	case "dotenv-linter", "shellcheck", "shfmt":
		return "shell"
	case "eslint":
		return "javascript"
	case "gitleaks":
		return "git"
	case "gofmt", "golangci-lint":
		return "go"
	case "google-java-format", "pmd":
		return "java"
	case "sqlfluff":
		return "sql"
	case "yamllint":
		return "yaml"
	case "qlty":
		return "custom"
	case "cppcheck":
		return "c"
	case "dotnetscanner":
		return "c#"
	default:
		return ""
	}
}

// MapLanguage 将 LANGUAGE_XXX 格式转为小写语言名
func MapLanguage(language string) string {
	if strings.HasPrefix(language, "LANGUAGE_") {
		lang := strings.TrimPrefix(language, "LANGUAGE_")
		return strings.ToLower(lang)
	}
	return "php"
}

// ScanTaskUploadReportReq 上传报告请求（从 multipart form 解析）
type ScanTaskUploadReportReq struct {
	TaskID  uint                `json:"taskID"`
	Method  constant.ScanMethod `json:"method"`
	Content []byte              `json:"-"`
}

// ScanTaskIssueReqFilter 任务问题列表查询过滤条件
type ScanTaskIssueReqFilter struct {
	TaskID          int64                    `json:"taskID" form:"taskID"`
	RepoID          int64                    `json:"repoID,omitempty" form:"repoID"`
	Branch          string                   `json:"branch,omitempty" form:"branch"`
	RuleID          int64                    `json:"ruleID,omitempty" form:"ruleID"`
	Status          constant.ScanIssueStatus `json:"status,omitempty" form:"status"`
	ScanMethod      constant.ScanMethod      `json:"scanMethod,omitempty" form:"scanMethod"`
	DetectionStatus string                   `json:"detectionStatus,omitempty" form:"detectionStatus"`
	Page            int                      `json:"page" form:"page" default:"1"`
	PageSize        int                      `json:"pageSize" form:"pageSize" default:"20"`
}

// ScanIssueReqFilter 任务问题列表查询过滤条件
type ScanIssueReqFilter struct {
	RepoID     int64                    `json:"repoID,omitempty" form:"repoID"`
	Branch     string                   `json:"branch,omitempty" form:"branch"`
	File       string                   `json:"file,omitempty" form:"file"`
	RuleID     int64                    `json:"ruleID,omitempty" form:"ruleID"`
	Status     constant.ScanIssueStatus `json:"status,omitempty" form:"status"`
	ScanMethod constant.ScanMethod      `json:"scanMethod,omitempty" form:"scanMethod"`
	Page       int                      `json:"page" form:"page" default:"1"`
	PageSize   int                      `json:"pageSize" form:"pageSize" default:"20"`
}

// ScanIssueIDsReqFilter 任务问题列表ID过滤条件
type ScanIssueIDsReqFilter struct {
	IDs []int64 `json:"ids" form:"ids[]" binding:"required"`
}

// ScanIssueUpdateStatusReq 批量更新扫描问题状态请求
type ScanIssueUpdateStatusReq struct {
	IDs        []int64                       `json:"ids" binding:"required"`
	Status     *constant.ScanIssueStatus     `json:"status"`
	Resolution *constant.ScanIssueResolution `json:"resolution"`
}

// ScanIssueInfo 扫描问题详情（含规则信息）
type ScanIssueInfo struct {
	ID              int64                         `json:"id"`
	Key             string                        `json:"key"`
	RuleID          int64                         `json:"ruleID"`
	RuleName        string                        `json:"ruleName"`
	RuleKey         string                        `json:"ruleKey"`
	RulePriority    constant.ScanRulePriority     `json:"rulePriority"`
	RuleType        constant.ScanRuleType         `json:"ruleType"`
	Message         string                        `json:"message"`
	Path            string                        `json:"path"`
	Line            int64                         `json:"line"`
	RepoID          int64                         `json:"repoID"`
	RepoBranch      string                        `json:"repoBranch"`
	Status          constant.ScanIssueStatus      `json:"status"`
	ScanMethod      constant.ScanMethod           `json:"scanMethod"`
	Resolution      *constant.ScanIssueResolution `json:"resolution"`
	Payload         json.RawMessage               `json:"payload"`
	CreatedByTaskID int64                         `json:"createdByTaskID"`
	UpdatedByTaskID int64                         `json:"updatedByTaskID"`
	CreatedDate     *datetime.Local               `json:"createdDate"`
}
