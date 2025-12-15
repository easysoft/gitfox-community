package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStructureValidation 测试结构验证功能
func TestStructureValidation(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("InvalidTypesJSON", func(t *testing.T) {
		// 创建无效的 types.json（不是数组）
		invalidTypes := `{"invalid": "structure"}`
		typesFile := filepath.Join(tmpDir, "invalid.types.json")
		rulesFile := filepath.Join(tmpDir, "invalid.rules.json")

		if err := os.WriteFile(typesFile, []byte(invalidTypes), 0644); err != nil {
			t.Fatal(err)
		}

		// 创建有效的 rules.json
		validRules := `{
  "version": "1.0",
  "source": "test",
  "lang": "test",
  "plugin": "test",
  "rules": [
    {
      "ruleKey": "rule1",
      "plugin": "test",
      "lang": "test",
      "name": "Test Rule",
      "priority": "high",
      "type": "defect",
      "tag": "test",
      "content": "test content",
      "content_en": "test content",
      "description": "test description",
      "description_en": "test description"
    }
  ]
}`
		if err := os.WriteFile(rulesFile, []byte(validRules), 0644); err != nil {
			t.Fatal(err)
		}

		result := ValidatePlugin("invalid", typesFile, rulesFile)

		if !result.HasStructureError {
			t.Error("Expected structure error for invalid types JSON")
		}

		if result.TypesStructureError == "" {
			t.Error("Expected types structure error message")
		}

		t.Logf("Types structure error: %s", result.TypesStructureError)
	})

	t.Run("EmptyTypesArray", func(t *testing.T) {
		// 创建空数组的 types.json
		emptyTypes := `[]`
		typesFile := filepath.Join(tmpDir, "empty.types.json")
		rulesFile := filepath.Join(tmpDir, "empty.rules.json")

		if err := os.WriteFile(typesFile, []byte(emptyTypes), 0644); err != nil {
			t.Fatal(err)
		}

		validRules := `{
  "version": "1.0",
  "source": "test",
  "lang": "test",
  "plugin": "test",
  "rules": []
}`
		if err := os.WriteFile(rulesFile, []byte(validRules), 0644); err != nil {
			t.Fatal(err)
		}

		result := ValidatePlugin("empty", typesFile, rulesFile)

		if !result.HasStructureError {
			t.Error("Expected structure error for empty types array")
		}

		t.Logf("Types structure error: %s", result.TypesStructureError)
	})

	t.Run("MissingRequiredRulesMetadata", func(t *testing.T) {
		// 创建有效的 types.json
		validTypes := `["rule1", "rule2"]`
		typesFile := filepath.Join(tmpDir, "meta.types.json")
		rulesFile := filepath.Join(tmpDir, "meta.rules.json")

		if err := os.WriteFile(typesFile, []byte(validTypes), 0644); err != nil {
			t.Fatal(err)
		}

		// 创建缺少必填字段的 rules.json
		invalidRules := `{
  "version": "",
  "source": "",
  "lang": "",
  "plugin": "",
  "rules": [
    {
      "ruleKey": "rule1",
      "plugin": "test",
      "lang": "test",
      "name": "Test Rule",
      "priority": "high",
      "type": "defect",
      "tag": "test",
      "content": "test content"
    }
  ]
}`
		if err := os.WriteFile(rulesFile, []byte(invalidRules), 0644); err != nil {
			t.Fatal(err)
		}

		result := ValidatePlugin("meta", typesFile, rulesFile)

		if !result.HasStructureError {
			t.Error("Expected structure error for missing metadata")
		}

		if result.RulesStructureError == "" {
			t.Error("Expected rules structure error message")
		}

		t.Logf("Rules structure error: %s", result.RulesStructureError)
	})

	t.Run("MissingRequiredRuleFields", func(t *testing.T) {
		// 创建有效的 types.json
		validTypes := `["rule1"]`
		typesFile := filepath.Join(tmpDir, "field.types.json")
		rulesFile := filepath.Join(tmpDir, "field.rules.json")

		if err := os.WriteFile(typesFile, []byte(validTypes), 0644); err != nil {
			t.Fatal(err)
		}

		// 创建缺少必填字段的规则
		invalidRules := `{
  "version": "1.0",
  "source": "test",
  "lang": "test",
  "plugin": "test",
  "rules": [
    {
      "ruleKey": "rule1",
      "plugin": "",
      "lang": "",
      "name": "",
      "priority": "",
      "type": "",
      "tag": "test"
    }
  ]
}`
		if err := os.WriteFile(rulesFile, []byte(invalidRules), 0644); err != nil {
			t.Fatal(err)
		}

		result := ValidatePlugin("field", typesFile, rulesFile)

		if !result.HasStructureError {
			t.Error("Expected structure error for missing rule fields")
		}

		t.Logf("Rules structure error: %s", result.RulesStructureError)
	})

	t.Run("ValidStructure", func(t *testing.T) {
		// 创建完全有效的文件
		validTypes := `["rule1", "rule2"]`
		typesFile := filepath.Join(tmpDir, "valid.types.json")
		rulesFile := filepath.Join(tmpDir, "valid.rules.json")

		if err := os.WriteFile(typesFile, []byte(validTypes), 0644); err != nil {
			t.Fatal(err)
		}

		validRules := `{
  "version": "1.0",
  "source": "test",
  "lang": "test",
  "plugin": "test",
  "rules": [
    {
      "ruleKey": "rule1",
      "plugin": "test",
      "lang": "test",
      "name": "Test Rule 1",
      "priority": "high",
      "type": "defect",
      "tag": "test",
      "content": "test content 1",
      "description": "test description 1"
    },
    {
      "ruleKey": "rule2",
      "plugin": "test",
      "lang": "test",
      "name": "Test Rule 2",
      "priority": "medium",
      "type": "security",
      "tag": "test",
      "content": "test content 2",
      "description": "test description 2"
    }
  ]
}`
		if err := os.WriteFile(rulesFile, []byte(validRules), 0644); err != nil {
			t.Fatal(err)
		}

		result := ValidatePlugin("valid", typesFile, rulesFile)

		if result.HasStructureError {
			t.Errorf("Expected no structure error, got: types=%s, rules=%s",
				result.TypesStructureError, result.RulesStructureError)
		}

		if !result.IsValid {
			t.Errorf("Expected validation to pass")
		}

		t.Log("✅ Valid structure passed all checks")
	})
}
