// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package main

import (
	"testing"
)

func TestValidateAllPlugins(t *testing.T) {
	// 使用当前目录进行测试
	results, err := ValidateAllPlugins(".")
	if err != nil {
		t.Fatalf("ValidateAllPlugins failed: %v", err)
	}

	// 打印验证结果
	PrintValidateResults(results)

	// 检查是否有验证失败的插件
	for _, result := range results {
		if result.ErrorMessage != "" {
			t.Errorf("Plugin %s has error: %s", result.Plugin, result.ErrorMessage)
			continue
		}

		if !result.IsValid {
			t.Errorf("Plugin %s validation failed. Missing keys: %v", result.Plugin, result.MissingKeys)
		}
	}
}

func TestValidateGolangciLint(t *testing.T) {
	result := ValidatePlugin("golangci-lint", "golangci-lint.types.json", "golangci-lint.rules.json")

	if result.ErrorMessage != "" {
		t.Fatalf("Error: %s", result.ErrorMessage)
	}

	t.Logf("Types Count: %d", result.TypesCount)
	t.Logf("Rules Count: %d", result.RulesCount)

	if !result.IsValid {
		t.Errorf("Validation failed. Missing keys: %v", result.MissingKeys)
	}

	if len(result.ExtraKeys) > 0 {
		t.Logf("Warning: Extra keys in rules: %v", result.ExtraKeys)
	}
}

func TestValidatePhpstan(t *testing.T) {
	result := ValidatePlugin("phpstan", "phpstan.types.json", "phpstan.rules.json")

	if result.ErrorMessage != "" {
		t.Fatalf("Error: %s", result.ErrorMessage)
	}

	t.Logf("Types Count: %d", result.TypesCount)
	t.Logf("Rules Count: %d", result.RulesCount)

	if !result.IsValid {
		t.Errorf("Validation failed. Missing keys: %v", result.MissingKeys)
	}

	if len(result.ExtraKeys) > 0 {
		t.Logf("Warning: Extra keys in rules: %v", result.ExtraKeys)
	}
}
