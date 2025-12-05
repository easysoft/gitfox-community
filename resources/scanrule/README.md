# Scan Rule Validator

扫描规则验证工具，用于检测 `types.json` 中的所有 key 在对应的 `rules.json` 中是否都存在。

## 目录结构

```
resources/scanrule/
├── golangci-lint.types.json    # golangci-lint 类型定义
├── golangci-lint.rules.json    # golangci-lint 规则定义
├── phpstan.types.json          # phpstan 类型定义
├── phpstan.rules.json          # phpstan 规则定义
├── validator.go                # 验证器核心代码
├── validator_test.go           # 测试用例
├── cmd/
│   └── validate/
│       └── main.go             # 命令行工具
└── README.md                   # 本文档
```

## 功能说明

该工具提供以下功能：

1. **验证所有插件**：检查目录下所有 `*.types.json` 和对应的 `*.rules.json` 文件
2. **验证单个插件**：指定插件名称，验证该插件的 types 和 rules 是否匹配
3. **结构验证**：验证 JSON 文件结构是否符合规范
   - Types 文件必须是非空的字符串数组
   - Rules 文件必须包含必填的元数据字段（version, source, lang, plugin）
   - 每个 rule 必须包含必填字段（ruleKey, plugin, lang, name, priority, type）
   - 每个 rule 至少要有 content 或 description
4. **重复检测**：检测 types.json 和 rules.json 中是否存在重复的 key
5. **详细报告**：输出验证结果，包括：
   - Types 数量
   - Rules 数量
   - 结构错误（如果存在）
   - 重复的 types（types 中重复出现的 key）
   - 重复的 rules（rules 中重复出现的 ruleKey）
   - 缺失的 keys（types 中有但 rules 中没有）
   - 额外的 keys（rules 中有但 types 中没有）

## 使用方法

### 方法一：使用 Task 命令（推荐）

#### 验证所有插件

```bash
task scanrule:check
```

#### 验证单个插件

```bash
task scanrule:check:plugin PLUGIN=golangci-lint
task scanrule:check:plugin PLUGIN=phpstan
```


### 方法二：使用命令行工具

#### 验证所有插件

```bash
cd resources/scanrule
go run cmd/validate/main.go -dir .
```

#### 验证单个插件

```bash
cd resources/scanrule
go run cmd/validate/main.go -dir . -plugin golangci-lint
```

#### 构建可执行文件

```bash
cd resources/scanrule/cmd/validate
go build -o scanrule-validator
./scanrule-validator -dir ../../
```

### 方法三：运行测试

```bash
cd resources/scanrule
go test -v
```

### 方法四：在代码中使用

```go
package main

import (
    "fmt"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/resources/scanrule"
)

func main() {
    // 验证所有插件
    results, err := scanrule.ValidateAllPlugins("./resources/scanrule")
    if err != nil {
        panic(err)
    }

    // 打印结果
    scanrule.PrintValidateResults(results)

    // 或者验证单个插件
    result := scanrule.ValidatePlugin(
        "golangci-lint",
        "./resources/scanrule/golangci-lint.types.json",
        "./resources/scanrule/golangci-lint.rules.json",
    )

    if !result.IsValid {
        fmt.Printf("Validation failed: %v\n", result.MissingKeys)
    }
}
```

## 输出示例

### 验证通过的情况

```
========================================
Scan Rule Validation Report
========================================

Plugin: golangci-lint
----------------------------------------
Types Count: 80
Rules Count: 80
✅ Validation: PASSED (All types have corresponding rules, no duplicates)

Plugin: phpstan
----------------------------------------
Types Count: 805
Rules Count: 805
✅ Validation: PASSED (All types have corresponding rules, no duplicates)

========================================
Summary: 2/2 plugins passed validation
========================================

✅ All validations passed!
```

### 验证失败的情况（缺失 key）

```
========================================
Scan Rule Validation Report
========================================

Plugin: golangci-lint
----------------------------------------
Types Count: 80
Rules Count: 78
❌ Validation: FAILED
   Missing Keys (2): [some-missing-key, another-missing-key]
⚠️  Extra Keys in Rules (0): []

========================================
Summary: 0/1 plugins passed validation
========================================
```

### 验证失败的情况（存在重复）

```
========================================
Scan Rule Validation Report
========================================

Plugin: test-plugin
----------------------------------------
Types Count: 5
Rules Count: 6
❌ Duplicate Types (2): [rule1 (x2), rule2 (x2)]
❌ Duplicate Rules (2): [rule1 (x2), rule4 (x2)]
❌ Validation: FAILED
   Has Duplicates: Yes
⚠️  Extra Keys in Rules (1): [rule4]

========================================
Summary: 0/1 plugins passed validation
========================================
```

### 验证失败的情况（结构错误）

```
========================================
Scan Rule Validation Report
========================================

Plugin: invalid-plugin
----------------------------------------
❌ Types Structure Error:
   types file must be a JSON array of strings: json: cannot unmarshal object into Go value of type []string
❌ Validation: FAILED (Structure errors detected)

========================================
Summary: 0/1 plugins passed validation
========================================
```

或者 rules 文件结构错误：

```
========================================
Scan Rule Validation Report
========================================

Plugin: invalid-rules
----------------------------------------
❌ Rules Structure Error:
   metadata validation errors: version field is required; source field is required; lang field is required; plugin field is required
❌ Validation: FAILED (Structure errors detected)

========================================
Summary: 0/1 plugins passed validation
========================================
```

## 数据结构

### types.json 格式

```json
[
  "asasalint",
  "bidichk",
  "copyloopvar",
  ...
]
```

### rules.json 格式

```json
{
  "version": "20250821",
  "source": "https://golangci-lint.run/usage/linters/",
  "lang": "go",
  "plugin": "golangci-lint",
  "rules": [
    {
      "ruleKey": "asasalint",          // 必填：规则唯一标识
      "plugin": "golangci-lint",       // 必填：插件名称
      "lang": "go",                    // 必填：编程语言
      "name": "规则名称",               // 必填：规则显示名称
      "priority": "high",              // 必填：优先级 (low/medium/high)
      "type": "defect",                // 必填：类型 (defect/security/compliance/optimize)
      "tag": "Best Practices",         // 可选：标签
      "content": "规则内容",            // content/content_en 至少一个
      "content_en": "Rule content",
      "description": "规则描述",        // description/description_en 至少一个
      "description_en": "Rule description"
    }
  ]
}
```

**必填字段说明：**
- **元数据字段**：`version`, `source`, `lang`, `plugin`
- **Rule 字段**：`ruleKey`, `plugin`, `lang`, `name`, `priority`, `type`
- **内容字段**：至少要有 `content`/`content_en` 或 `description`/`description_en` 之一

## 验证逻辑

1. **读取文件**：读取 `*.types.json` 和 `*.rules.json` 文件
2. **结构验证**：
   - 验证 types 文件是否为非空字符串数组
   - 验证 rules 文件是否包含所有必填字段
   - 验证每个 rule 是否包含所有必填字段
3. **重复检测**：统计 types 和 rules 中重复出现的 key
4. **匹配检查**：检查每个 types 中的 key 是否在 rules 的 ruleKey 中存在
5. **额外 key 检查**：检查 rules 中是否有 types 中不存在的 key（警告信息）
6. **验证通过条件**：
   - 文件结构正确
   - 所有 types 中的 key 都在 rules 中存在
   - types 中没有重复的 key
   - rules 中没有重复的 ruleKey

## API 参考

### ValidateAllPlugins

```go
func ValidateAllPlugins(dir string) ([]ValidateResult, error)
```

验证指定目录下所有插件的 types 和 rules 是否匹配。

**参数：**
- `dir`: 包含规则文件的目录路径

**返回：**
- `[]ValidateResult`: 所有插件的验证结果
- `error`: 错误信息

### ValidatePlugin

```go
func ValidatePlugin(plugin, typesFile, rulesFile string) ValidateResult
```

验证单个插件的 types 和 rules 是否匹配。

**参数：**
- `plugin`: 插件名称
- `typesFile`: types.json 文件路径
- `rulesFile`: rules.json 文件路径

**返回：**
- `ValidateResult`: 验证结果

### ValidateResult 结构

```go
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
    IsValid             bool     // 是否所有验证都通过
    HasDuplicates       bool     // 是否存在重复的 key
    HasStructureError   bool     // 是否存在结构错误
    ErrorMessage        string   // 错误信息
}
```


## 维护指南

### 添加新的插件规则

1. 创建 `<plugin-name>.types.json` 文件，包含所有类型的 key 列表
2. 创建 `<plugin-name>.rules.json` 文件，包含所有规则的详细定义
3. 运行验证工具确保所有 types 中的 key 都有对应的 rule

### 更新现有规则

1. 修改 types.json 或 rules.json 文件
2. 运行 `go test -v` 或 `go run cmd/validate/main.go` 验证更改
3. 确保验证通过后再提交

## 常见问题

**Q: 为什么会出现 Extra Keys 警告？**

A: Extra Keys 表示 rules.json 中有一些 ruleKey，但在 types.json 中没有定义。这通常不是错误，可能是规则文件包含了一些额外的、可选的规则。

**Q: 如何修复 Missing Keys 错误？**

A: Missing Keys 表示 types.json 中定义的某些 key 在 rules.json 中找不到对应的规则。需要在 rules.json 中添加这些缺失的规则定义。

**Q: 如何修复 Duplicate Keys 错误？**

A: Duplicate Keys 表示在 types.json 或 rules.json 中有重复的 key。需要手动编辑文件，删除重复的条目。验证工具会显示哪些 key 重复了以及重复的次数（例如：`rule1 (x2)` 表示 rule1 出现了 2 次）。

**Q: 如何修复 Structure Errors？**

A: Structure Errors 表示文件结构不符合规范。根据错误信息修复：
- **Types 文件**：必须是一个字符串数组，例如 `["rule1", "rule2"]`
- **Rules 文件**：必须包含 `version`, `source`, `lang`, `plugin` 和 `rules` 字段
- **Rule 条目**：每个 rule 必须包含 `ruleKey`, `plugin`, `lang`, `name`, `priority`, `type` 字段，以及至少一个 `content`/`description` 字段

**Q: 验证工具会修改文件吗？**

A: 不会。验证工具只读取文件内容进行检查，不会修改任何文件。
