package scan

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/bmatcuk/doublestar/v4"
)

// Pattern 扫描计划分支/文件匹配模式
// 结构与参考项目 protection.Pattern 保持一致。
type Pattern struct {
	Default bool     `json:"default,omitempty"`
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// JSONString 将 Pattern 转换为 JSON 字符串
func (p Pattern) JSONString() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParsePattern 解析 JSON 字符串为 Pattern 结构体
func ParsePattern(raw string) (*Pattern, error) {
	if raw == "" {
		return &Pattern{}, nil
	}

	var pattern Pattern
	if err := json.Unmarshal([]byte(raw), &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

// Matches 检查分支是否匹配 Pattern
func (p Pattern) Matches(name, defaultName string) bool {
	matches := !p.Default && len(p.Include) == 0
	matches = matches || p.Default && name == defaultName

	if !matches {
		for _, include := range p.Include {
			if matches = patternMatches(include, name); matches {
				break
			}
		}
	}

	for _, exclude := range p.Exclude {
		matches = matches && !patternMatches(exclude, name)
	}

	return matches
}

// Validate 验证 Pattern 是否有效
func (p Pattern) Validate(field string) error {
	if err := p.validatePatterns(field, "include", p.Include); err != nil {
		return err
	}
	if err := p.validatePatterns(field, "exclude", p.Exclude); err != nil {
		return err
	}
	return nil
}

// validatePatterns 验证 Pattern 中的每个模式是否有效
func (p Pattern) validatePatterns(field string, patternType string, patterns []string) error {
	for _, pattern := range patterns {
		if err := validatePatternValue(pattern); err != nil {
			return wrapPatternError(field, patternType, err)
		}
	}
	return nil
}

// validatePatternValue 验证单个模式是否有效
func validatePatternValue(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	if _, err := doublestar.Match(pattern, "test"); err != nil {
		return err
	}
	return nil
}

// patternMatches 检查分支/文件名是否匹配模式
func patternMatches(pattern, name string) bool {
	ok, _ := doublestar.Match(pattern, name)
	return ok
}

// wrapPatternError 包装 Pattern 错误信息
func wrapPatternError(field string, patternType string, err error) error {
	if err.Error() == "pattern cannot be empty" {
		return errors.WithI18nKey(err, i18n.PlanPatternEmptyError)
	}

	key := i18n.PlanPatternEmptyError
	switch strings.ToLower(field) {
	case "branches":
		switch patternType {
		case "include":
			key = i18n.PlanBranchesIncludeError
		case "exclude":
			key = i18n.PlanBranchesExcludeError
		}
	case "files":
		switch patternType {
		case "include":
			key = i18n.PlanFilesIncludeError
		case "exclude":
			key = i18n.PlanFilesExcludeError
		}
	}
	return errors.WithI18nKey(fmt.Errorf("invalid %s %s pattern: %w", field, patternType, err), key)
}

// ScanPlanReqForCreate 创建扫描计划请求
type ScanPlanReqForCreate struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Desc        string  `json:"desc"`
	SolutionIDs []uint  `json:"solutionIDs"`
	ScanType    string  `json:"scanType"`
	Branches    Pattern `json:"branches"`
	Files       Pattern `json:"files"`
}

// ScanPlanReqForUpdate 更新扫描计划请求
type ScanPlanReqForUpdate struct {
	Name     string   `json:"name,omitempty"`
	Desc     string   `json:"desc,omitempty"`
	ScanType string   `json:"scanType,omitempty"`
	Branches *Pattern `json:"branches,omitempty"`
	Files    *Pattern `json:"files,omitempty"`
}

// ScanPlanReqFilter 查询扫描计划过滤条件
type ScanPlanReqFilter struct {
	Name             string `json:"name"`
	RepoID           string `json:"repo_id,omitempty"`
	SpaceID          string `json:"space_id,omitempty"`
	LatestTaskResult string `json:"latest_task_result,omitempty"`
	LatestTaskStatus string `json:"latest_task_status,omitempty"`
	Sort             string `json:"sort" form:"sort" default:"id"`
	Order            string `json:"order" form:"order" default:"desc"`
	Page             int    `json:"page" form:"page" default:"1"`
	PageSize         int    `json:"pageSize" form:"pageSize" default:"30"`
}

// ScanPlanReqForBind 绑定/解绑扫描方案请求
type ScanPlanReqForBind struct {
	SolutionIDs []uint `json:"solutionIDs"`
}

// ScanPlanRulesReqQuery 扫描计划规则查询参数
type ScanPlanRulesReqQuery struct {
	TriggerID uint `form:"triggerID"`
}

// ScanPlan 扫描计划响应模型
type ScanPlan struct {
	ID            int64       `json:"id"`
	RepoID        int64       `json:"repoID"`
	Name          string      `json:"name"`
	Desc          string      `json:"desc"`
	ScanType      string      `json:"scanType"`
	Branches      *Pattern    `json:"branches,omitempty"`
	Files         *Pattern    `json:"files,omitempty"`
	SolutionIDs   []int64     `json:"solutionIDs,omitempty"`
	SolutionCount int         `json:"solutionCount"`
	TriggerCount  int         `json:"triggerCount"`
	Solutions     []*Solution `json:"solutions,omitempty"`
	CreatedBy     string      `json:"createdBy,omitempty"`
	CreatedDate   *time.Time  `json:"createdDate"`
	EditedBy      string      `json:"editedBy"`
	EditedDate    *time.Time  `json:"editedDate"`
	Deleted       bool        `json:"deleted"`
}
