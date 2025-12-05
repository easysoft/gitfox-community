package scanrule

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RulesFile 规则文件结构
type RulesFile struct {
	Version string `json:"version"`
	Source  string `json:"source"`
	Lang    string `json:"lang"`
	Plugin  string `json:"plugin"`
	Rules   []Rule `json:"rules"`
}

// Rule 单个规则结构
type Rule struct {
	RuleKey     string `json:"ruleKey"`
	Plugin      string `json:"plugin"`
	Lang        string `json:"lang"`
	Name        string `json:"name"`
	Priority    string `json:"priority"`
	Type        string `json:"type"`
	Tag         string `json:"tag"`
	Content     string `json:"content"`
	ContentEn   string `json:"content_en"`
	Description string `json:"description"`
	DescEn      string `json:"description_en"`
}

// ValidateResult 验证结果
type ValidateResult struct {
	Plugin              string   // 插件名称
	TypesCount          int      // types 文件中的 key 数量
	RulesCount          int      // rules 文件中的 key 数量
	MissingKeys         []string // types 中有但 rules 中没有的 key
	ExtraKeys           []string // rules 中有但 types 中没有的 key
	DuplicateTypes      []string // types 中重复的 key
	DuplicateRules      []string // rules 中重复的 ruleKey
	TypesStructureError string   // types 文件结构错误信息
	RulesStructureError string   // rules 文件结构错误信息
	IsValid             bool     // 是否所有 types 中的 key 都在 rules 中
	HasDuplicates       bool     // 是否存在重复的 key
	HasStructureError   bool     // 是否存在结构错误
	ErrorMessage        string   // 错误信息
}

// ValidateAllPlugins 验证指定目录下所有插件的 types 和 rules 是否匹配
func ValidateAllPlugins(dir string) ([]ValidateResult, error) {
	// 获取目录下所有 types 文件
	typesFiles, err := filepath.Glob(filepath.Join(dir, "*.types.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob types files: %w", err)
	}

	var results []ValidateResult

	for _, typesFile := range typesFiles {
		// 从文件名提取插件名称（例如：golangci-lint.types.json -> golangci-lint）
		plugin := strings.TrimSuffix(filepath.Base(typesFile), ".types.json")

		// 构造对应的 rules 文件路径
		rulesFile := filepath.Join(dir, plugin+".rules.json")

		// 验证这对文件
		result := ValidatePlugin(plugin, typesFile, rulesFile)
		results = append(results, result)
	}

	return results, nil
}

// ValidatePlugin 验证单个插件的 types 和 rules 是否匹配
func ValidatePlugin(plugin, typesFile, rulesFile string) ValidateResult {
	result := ValidateResult{
		Plugin:         plugin,
		MissingKeys:    []string{},
		ExtraKeys:      []string{},
		DuplicateTypes: []string{},
		DuplicateRules: []string{},
	}

	// 读取并验证 types 文件结构
	types, typesErr := readAndValidateTypesFile(typesFile)
	if typesErr != nil {
		result.TypesStructureError = typesErr.Error()
		result.HasStructureError = true
	} else {
		result.TypesCount = len(types)
	}

	// 读取并验证 rules 文件结构
	rulesFileData, ruleKeys, rulesErr := readAndValidateRulesFile(rulesFile)
	if rulesErr != nil {
		result.RulesStructureError = rulesErr.Error()
		result.HasStructureError = true
	} else {
		result.RulesCount = len(ruleKeys)

		// 验证 rules 文件的元数据字段
		if err := validateRulesMetadata(rulesFileData); err != nil {
			result.RulesStructureError = err.Error()
			result.HasStructureError = true
		}

		// 验证每个 rule 的必填字段
		if err := validateRulesFields(rulesFileData.Rules); err != nil {
			result.RulesStructureError = err.Error()
			result.HasStructureError = true
		}
	}

	// 如果有结构错误，标记为验证失败并返回
	if result.HasStructureError {
		result.IsValid = false
		return result
	}

	// 检查 types 中的重复项
	typesSet := make(map[string]int)
	for _, key := range types {
		typesSet[key]++
	}
	for key, count := range typesSet {
		if count > 1 {
			result.DuplicateTypes = append(result.DuplicateTypes, fmt.Sprintf("%s (x%d)", key, count))
		}
	}

	// 检查 rules 中的重复项
	ruleKeySet := make(map[string]int)
	for _, key := range ruleKeys {
		ruleKeySet[key]++
	}
	for key, count := range ruleKeySet {
		if count > 1 {
			result.DuplicateRules = append(result.DuplicateRules, fmt.Sprintf("%s (x%d)", key, count))
		}
	}

	// 标记是否存在重复
	result.HasDuplicates = len(result.DuplicateTypes) > 0 || len(result.DuplicateRules) > 0

	// 检查 types 中的每个 key 是否在 rules 中存在（使用去重后的集合）
	for key := range typesSet {
		if ruleKeySet[key] == 0 {
			result.MissingKeys = append(result.MissingKeys, key)
		}
	}

	// 检查 rules 中是否有 types 中没有的 key
	for key := range ruleKeySet {
		if typesSet[key] == 0 {
			result.ExtraKeys = append(result.ExtraKeys, key)
		}
	}

	// 如果没有缺失的 key 且没有重复且没有结构错误，则验证通过
	result.IsValid = len(result.MissingKeys) == 0 && !result.HasDuplicates && !result.HasStructureError

	return result
}

// readTypesFile 读取 types 文件，返回 key 列表
func readTypesFile(filepath string) ([]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var types []string
	if err := json.Unmarshal(data, &types); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return types, nil
}

// readRulesFile 读取 rules 文件，返回所有 ruleKey 列表
func readRulesFile(filepath string) ([]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var rulesFile RulesFile
	if err := json.Unmarshal(data, &rulesFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	ruleKeys := make([]string, 0, len(rulesFile.Rules))
	for _, rule := range rulesFile.Rules {
		ruleKeys = append(ruleKeys, rule.RuleKey)
	}

	return ruleKeys, nil
}

// readAndValidateTypesFile 读取并验证 types 文件结构
func readAndValidateTypesFile(filepath string) ([]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 首先检查是否为有效的 JSON
	var rawJSON interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	// 检查是否为数组
	var types []string
	if err := json.Unmarshal(data, &types); err != nil {
		return nil, fmt.Errorf("types file must be a JSON array of strings: %w", err)
	}

	// 验证数组不为空
	if len(types) == 0 {
		return nil, fmt.Errorf("types array cannot be empty")
	}

	// 验证每个元素都是非空字符串
	for i, t := range types {
		if strings.TrimSpace(t) == "" {
			return nil, fmt.Errorf("types[%d] is empty or contains only whitespace", i)
		}
	}

	return types, nil
}

// readAndValidateRulesFile 读取并验证 rules 文件结构
func readAndValidateRulesFile(filepath string) (*RulesFile, []string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 首先检查是否为有效的 JSON
	var rawJSON interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	var rulesFile RulesFile
	if err := json.Unmarshal(data, &rulesFile); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal rules file: %w", err)
	}

	ruleKeys := make([]string, 0, len(rulesFile.Rules))
	for _, rule := range rulesFile.Rules {
		ruleKeys = append(ruleKeys, rule.RuleKey)
	}

	return &rulesFile, ruleKeys, nil
}

// validateRulesMetadata 验证 rules 文件的元数据字段
func validateRulesMetadata(rulesFile *RulesFile) error {
	var errors []string

	if strings.TrimSpace(rulesFile.Version) == "" {
		errors = append(errors, "version field is required")
	}

	if strings.TrimSpace(rulesFile.Source) == "" {
		errors = append(errors, "source field is required")
	}

	if strings.TrimSpace(rulesFile.Lang) == "" {
		errors = append(errors, "lang field is required")
	}

	if strings.TrimSpace(rulesFile.Plugin) == "" {
		errors = append(errors, "plugin field is required")
	}

	if len(rulesFile.Rules) == 0 {
		errors = append(errors, "rules array cannot be empty")
	}

	if len(errors) > 0 {
		return fmt.Errorf("metadata validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateRulesFields 验证每个 rule 的必填字段
func validateRulesFields(rules []Rule) error {
	for i, rule := range rules {
		var errors []string

		if strings.TrimSpace(rule.RuleKey) == "" {
			errors = append(errors, "ruleKey is required")
		}

		if strings.TrimSpace(rule.Plugin) == "" {
			errors = append(errors, "plugin is required")
		}

		if strings.TrimSpace(rule.Lang) == "" {
			errors = append(errors, "lang is required")
		}

		if strings.TrimSpace(rule.Name) == "" {
			errors = append(errors, "name is required")
		}

		if strings.TrimSpace(rule.Priority) == "" {
			errors = append(errors, "priority is required")
		}

		if strings.TrimSpace(rule.Type) == "" {
			errors = append(errors, "type is required")
		}

		// content 和 description 至少要有一个
		hasContent := strings.TrimSpace(rule.Content) != "" || strings.TrimSpace(rule.ContentEn) != ""
		hasDescription := strings.TrimSpace(rule.Description) != "" || strings.TrimSpace(rule.DescEn) != ""

		if !hasContent && !hasDescription {
			errors = append(errors, "at least one of content/content_en or description/description_en is required")
		}

		if len(errors) > 0 {
			return fmt.Errorf("rule[%d] (ruleKey=%s) validation errors: %s",
				i, rule.RuleKey, strings.Join(errors, "; "))
		}
	}

	return nil
}

// PrintValidateResults 打印验证结果
func PrintValidateResults(results []ValidateResult) {
	fmt.Println("========================================")
	fmt.Println("Scan Rule Validation Report")
	fmt.Println("========================================")

	for _, result := range results {
		fmt.Printf("\nPlugin: %s\n", result.Plugin)
		fmt.Println(strings.Repeat("-", 40))

		if result.ErrorMessage != "" {
			fmt.Printf("❌ Error: %s\n", result.ErrorMessage)
			continue
		}

		// 检查结构错误
		if result.HasStructureError {
			if result.TypesStructureError != "" {
				fmt.Printf("❌ Types Structure Error:\n   %s\n", result.TypesStructureError)
			}
			if result.RulesStructureError != "" {
				fmt.Printf("❌ Rules Structure Error:\n   %s\n", result.RulesStructureError)
			}
			fmt.Printf("❌ Validation: FAILED (Structure errors detected)\n")
			continue
		}

		fmt.Printf("Types Count: %d\n", result.TypesCount)
		fmt.Printf("Rules Count: %d\n", result.RulesCount)

		// 检查重复项
		if len(result.DuplicateTypes) > 0 {
			fmt.Printf("❌ Duplicate Types (%d): %v\n", len(result.DuplicateTypes), result.DuplicateTypes)
		}

		if len(result.DuplicateRules) > 0 {
			fmt.Printf("❌ Duplicate Rules (%d): %v\n", len(result.DuplicateRules), result.DuplicateRules)
		}

		// 检查缺失和额外的 key
		if result.IsValid {
			fmt.Printf("✅ Validation: PASSED (All types have corresponding rules, no duplicates, structure valid)\n")
		} else {
			fmt.Printf("❌ Validation: FAILED\n")
			if len(result.MissingKeys) > 0 {
				fmt.Printf("   Missing Keys (%d): %v\n", len(result.MissingKeys), result.MissingKeys)
			}
			if result.HasDuplicates {
				fmt.Printf("   Has Duplicates: Yes\n")
			}
		}

		if len(result.ExtraKeys) > 0 {
			fmt.Printf("⚠️  Extra Keys in Rules (%d): %v\n", len(result.ExtraKeys), result.ExtraKeys)
		}
	}

	fmt.Println("\n========================================")

	// 统计总体结果
	totalPlugins := len(results)
	passedPlugins := 0
	for _, result := range results {
		if result.IsValid && result.ErrorMessage == "" && !result.HasStructureError {
			passedPlugins++
		}
	}

	fmt.Printf("Summary: %d/%d plugins passed validation\n", passedPlugins, totalPlugins)
	fmt.Println("========================================")
}
