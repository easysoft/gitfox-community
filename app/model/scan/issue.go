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
	Key             string                        `json:"key"`
	RuleID          int64                         `json:"ruleID"`
	PpmID           int64                         `json:"ppmID,omitempty"`
	Title           string                        `json:"title"`
	Content         string                        `json:"content,omitempty"`
	Path            string                        `json:"path"`
	StartLine       int64                         `json:"startLine"`
	EndLine         int64                         `json:"endLine"`
	OldCode         string                        `json:"oldCode,omitempty"`
	NewCode         string                        `json:"newCode,omitempty"`
	RepoID          int64                         `json:"repoID"`
	RepoBranch      string                        `json:"repoBranch"`
	CreatedByTaskID int64                         `json:"createdByTaskID"`
	UpdatedByTaskID int64                         `json:"updatedByTaskID"`
	Status          constant.ScanIssueStatus      `json:"status"`
	ScanMethod      constant.ScanMethod           `json:"scanMethod"`
	Resolution      *constant.ScanIssueResolution `json:"resolution"`
	Category        constant.AiScanIssueCategory  `json:"category,omitempty"`
	Severity        constant.AiScanIssueSeverity  `json:"severity,omitempty"`
	Payload         json.RawMessage               `json:"payload"`
	CreatedBy       string                        `json:"createdBy,omitempty"`
	EditedBy        string                        `json:"editedBy,omitempty"`
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
	IgnoredDate     *int64                   `json:"ignoredDate"`
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

// AIReportResponse AI 审查报告顶层结构
type AIReportResponse struct {
	Status   string           `json:"status"`
	Summary  *AIReportSummary `json:"summary"`
	Comments []*IssueInputAI  `json:"comments"`
}

// AIReportSummary AI 审查报告摘要
type AIReportSummary struct {
	FilesReviewed int    `json:"files_reviewed"`
	Comments      int    `json:"comments"`
	TotalTokens   int64  `json:"total_tokens"`
	InputTokens   int64  `json:"input_tokens"`
	OutputTokens  int64  `json:"output_tokens"`
	Elapsed       string `json:"elapsed"`
}

// IssueInputAI AI 类型的 issue 输入
type IssueInputAI struct {
	Path           string `json:"path"`
	Content        string `json:"content"`
	SuggestionCode string `json:"suggestion_code"`
	ExistingCode   string `json:"existing_code"`
	StartLine      int64  `json:"start_line"`
	EndLine        int64  `json:"end_line"`
	Category       string `json:"category"`
	Severity       string `json:"severity"`
}

// IssuePayloadAI AI 类型 issue 的扩展数据
type IssuePayloadAI struct {
	SuggestionCode string `json:"suggestion_code"`
	ExistingCode   string `json:"existing_code"`
}

// UniqueKey 生成唯一标识
func (t IssueInputAI) UniqueKey() string {
	return fmt.Sprintf("ai/%s/%d/%d", t.Path, t.StartLine, t.EndLine)
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
	ID              int64                        `json:"id,omitempty" form:"id"`
	Title           string                       `json:"title,omitempty" form:"title"`
	Content         string                       `json:"content,omitempty" form:"content"`
	Path            string                       `json:"path,omitempty" form:"path"`
	RuleType        constant.ScanRuleType        `json:"ruleType,omitempty" form:"ruleType"`
	Tool            string                       `json:"tool,omitempty" form:"tool"`
	RulePriority    string                       `json:"rulePriority,omitempty" form:"rulePriority"`
	CreateDate      string                       `json:"createDate,omitempty" form:"createDate"`
	TaskID          int64                        `json:"taskID" form:"taskID"`
	RepoID          int64                        `json:"repoID,omitempty" form:"repoID"`
	PpmID           int64                        `json:"ppmID,omitempty" form:"ppmID"`
	Category        constant.AiScanIssueCategory `json:"category,omitempty" form:"category"`
	Severity        constant.AiScanIssueSeverity `json:"severity,omitempty" form:"severity"`
	Branch          string                       `json:"branch,omitempty" form:"branch"`
	RuleID          int64                        `json:"ruleID,omitempty" form:"ruleID"`
	Status          constant.ScanIssueStatus     `json:"status,omitempty" form:"status"`
	ScanMethods     []constant.ScanMethod        `json:"scanMethods,omitempty" form:"scanMethods"`
	PlanID          int64                        `json:"planID,omitempty" form:"planID"`
	DetectionStatus string                       `json:"detectionStatus,omitempty" form:"detectionStatus"`
	Sort            string                       `json:"sort,omitempty" form:"sort" default:"id"`
	Order           string                       `json:"order,omitempty" form:"order" default:"desc"`
	Page            int                          `json:"page" form:"page" default:"1"`
	PageSize        int                          `json:"pageSize" form:"pageSize" default:"20"`
}

// ScanIssueReqFilter 任务问题列表查询过滤条件
type ScanIssueReqFilter struct {
	ID           int64                        `json:"id,omitempty" form:"id"`
	RepoID       int64                        `json:"repoID,omitempty" form:"repoID"`
	PpmID        int64                        `json:"ppmID,omitempty" form:"ppmID"`
	Category     constant.AiScanIssueCategory `json:"category,omitempty" form:"category"`
	Severity     constant.AiScanIssueSeverity `json:"severity,omitempty" form:"severity"`
	Branch       string                       `json:"branch,omitempty" form:"branch"`
	Path         string                       `json:"path,omitempty" form:"path"`
	Title        string                       `json:"title,omitempty" form:"title"`
	Content      string                       `json:"content,omitempty" form:"content"`
	RuleID       int64                        `json:"ruleID,omitempty" form:"ruleID"`
	RulePriority string                       `json:"rulePriority,omitempty" form:"rulePriority"`
	RuleType     constant.ScanRuleType        `json:"ruleType,omitempty" form:"ruleType"`
	Tool         string                       `json:"tool,omitempty" form:"tool"`
	Status       constant.ScanIssueStatus     `json:"status,omitempty" form:"status"`
	ScanMethods  []constant.ScanMethod        `json:"scanMethods,omitempty" form:"scanMethods"`
	CreateDate   string                       `json:"createDate,omitempty" form:"createDate"`
	CreatedFilter
	UpdatedFilter
	ResolvedFilter
	PlanID   int64  `json:"planID,omitempty" form:"planID"`
	Sort     string `json:"sort,omitempty" form:"sort" default:"id"`
	Order    string `json:"order,omitempty" form:"order" default:"desc"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"20"`
}

type CreatedFilter struct {
	CreatedGt string `json:"createdGt,omitempty" form:"createdGt"`
	CreatedLt string `json:"createdLt,omitempty" form:"createdLt"`
}

type UpdatedFilter struct {
	UpdatedGt string `json:"updatedGt,omitempty" form:"updatedGt"`
	UpdatedLt string `json:"updatedLt,omitempty" form:"updatedLt"`
}

type ResolvedFilter struct {
	ResolvedGt string `json:"resolvedGt,omitempty" form:"resolvedGt"`
	ResolvedLt string `json:"resolvedLt,omitempty" form:"resolvedLt"`
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
	Resolved   *datetime.Local               `json:"resolved"`
	Ignored    *int64                        `json:"ignored"`
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
	Title           string                        `json:"title"`
	Content         string                        `json:"content,omitempty"`
	Path            string                        `json:"path"`
	StartLine       int64                         `json:"startLine"`
	EndLine         int64                         `json:"endLine"`
	OldCode         string                        `json:"oldCode,omitempty"`
	NewCode         string                        `json:"newCode,omitempty"`
	Category        constant.AiScanIssueCategory  `json:"category,omitempty"`
	Severity        constant.AiScanIssueSeverity  `json:"severity,omitempty"`
	PpmID           int64                         `json:"ppmID,omitempty"`
	PpmTitle        string                        `json:"ppmTitle,omitempty"`
	RepoID          int64                         `json:"repoID"`
	RepoBranch      string                        `json:"repoBranch"`
	Status          constant.ScanIssueStatus      `json:"status"`
	ScanMethod      constant.ScanMethod           `json:"scanMethod"`
	Resolution      *constant.ScanIssueResolution `json:"resolution"`
	Payload         json.RawMessage               `json:"payload"`
	CreatedByTaskID int64                         `json:"createdByTaskID"`
	UpdatedByTaskID int64                         `json:"updatedByTaskID"`
	PlanID          int64                         `json:"planID"`
	PlanName        string                        `json:"planName"`
	CreatedDate     *datetime.Local               `json:"createdDate"`
	ResolvedDate    *datetime.Local               `json:"resolvedDate"`
	ClosedDate      *datetime.Local               `json:"closedDate"`
	IgnoredDate     *int64                        `json:"ignoredDate"`
	CreatedBy       string                        `json:"createdBy,omitempty"`
	EditedBy        string                        `json:"editedBy,omitempty"`
	Trigger         *ScanTaskTrigger              `json:"trigger"`
}
