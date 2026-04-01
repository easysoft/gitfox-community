# CLI 多语言支持

GitFox CLI 支持中英文双语，会根据系统环境自动检测并显示对应语言的帮助信息。

## 支持的语言

- `zh-CN`: 简体中文
- `en-US`: 英文（默认）

## 语言检测优先级

CLI 会按以下优先级检测语言环境：

1. `GITFOX_LANG` 环境变量（最高优先级）
2. `LANG` 环境变量
3. `LC_ALL` 环境变量
4. 默认英文

## 使用示例

### 使用默认语言（英文）

```bash
./gitfox --help
```

### 使用中文

```bash
# 方式 1: 使用 GITFOX_LANG 环境变量
GITFOX_LANG=zh-CN ./gitfox --help

# 方式 2: 使用 LANG 环境变量
LANG=zh_CN.UTF-8 ./gitfox --help

# 方式 3: 永久设置（添加到 ~/.bashrc 或 ~/.zshrc）
export GITFOX_LANG=zh-CN
./gitfox --help
```

### 使用英文

```bash
GITFOX_LANG=en-US ./gitfox --help
```

## 添加新的翻译

CLI 多语言使用项目统一的 i18n 系统，翻译文件位于 `resources/locales/` 目录。

### 步骤 1: 编辑 YAML 文件

编辑对应语言的 `cli.yaml` 文件：

**resources/locales/en-US/cli.yaml**
```yaml
cli:
  server: "server"
  server_description: "starts the server"
  server_mode_arg: "set the server mode"
  server_enable_ci_flag: "start ci runners for build executions"
  # 添加新的翻译键...
```

**resources/locales/zh-CN/cli.yaml**
```yaml
cli:
  server: "server"
  server_description: "启动服务器"
  server_mode_arg: "设置服务器模式"
  server_enable_ci_flag: "启动 CI 运行器以执行构建"
  # 添加新的翻译键...
```

### 步骤 2: 生成常量

运行 task 命令生成 Go 常量：

```bash
task i18n:generate
```

这会自动更新 `pkg/i18n/i18n_keys.go` 文件，生成类似 `CliServerDescription` 的常量。

### 步骤 3: 在代码中使用

在命令注册代码中使用生成的常量：

```go
import (
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
    clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
)

func Register(app *kingpin.Application) {
    cmd := app.Command(
        clii18n.T(i18n.CliServer),
        clii18n.T(i18n.CliServerDescription),
    )
}
```

## 实现原理

1. CLI 启动时调用 `clii18n.Init()` 初始化多语言
2. **直接从 embedded resources 读取翻译文件到内存**（无需写入文件系统）
3. 使用 `embed.FS` 的 `ReadFile` 方法读取 YAML/JSON 文件内容
4. 通过 i18n Loader 的 `LoadFromBytes` 方法解析翻译
5. 将解析后的翻译直接加载到 i18n Manager 的内存中
6. 根据环境变量自动检测语言环境
7. 各个命令注册时调用 `clii18n.T(key)` 获取对应语言的文本
8. kingpin 使用这些文本生成帮助信息

### 性能优化

**零文件系统 I/O 设计：**
- ✅ **无临时文件**：翻译直接从二进制文件的 embedded 区域读取到内存
- ✅ **无磁盘写入**：不创建任何临时目录或文件
- ✅ **快速启动**：避免了文件复制和磁盘 I/O 开销
- ✅ **跨平台一致**：不依赖文件系统权限和路径处理
- ✅ **多实例安全**：无临时文件，不存在冲突问题

**内存加载流程：**
```
二进制文件 (embedded) → embed.FS.ReadFile("locales/{locale}/cli.yaml") → 
内存字节数组 → Loader.LoadFromBytes() → 解析为 map[string]string → 
Manager.AddTranslation() → 内存中的翻译表
```

**加载策略：**
- 只加载 `cli.yaml` 文件，不加载其他翻译文件
- 支持的语言：`en-US`, `zh-CN`（硬编码在代码中）
- 如果某个语言的 `cli.yaml` 不存在，会跳过该语言
- 至少需要一个语言的翻译文件才能成功初始化

**优势对比：**
| 方案 | 文件 I/O | 启动速度 | 临时文件 | 多实例安全 |
|------|---------|---------|---------|-----------|
| 旧方案（复制到临时目录） | 多次读写 | 慢 | 需要清理 | 需要 PID+时间戳 |
| 新方案（内存加载） | 零 I/O | 快 | 无 | 天然安全 |

## 架构优势

- 使用项目统一的 i18n 系统，保持一致性
- 翻译文件集中管理在 `resources/locales/`
- 通过 task 自动生成类型安全的常量
- 支持嵌套的 YAML 结构，便于组织
- 自动进行键一致性校验
- **零文件系统 I/O**：翻译直接从二进制 embedded 区域加载到内存
- **快速启动**：无需创建临时文件，启动速度更快
- **无需清理**：不产生临时文件，无需清理逻辑

## 技术细节

### Embedded Resources

翻译文件通过 Go 1.16+ 的 `embed` 包嵌入到二进制文件中：

```go
//go:embed locales
var EmbeddedLocales embed.FS
```

编译后，所有 `resources/locales/` 下的文件都会被打包到可执行文件中，无需单独分发。

### 内存加载机制

```go
// 1. 遍历支持的语言列表
locales := []string{"en-US", "zh-CN"}

// 2. 读取每个语言的 cli.yaml 到内存
data, _ := resources.EmbeddedLocales.ReadFile("locales/zh-CN/cli.yaml")

// 3. 解析为翻译 map
translations, _ := loader.LoadFromBytes(data, ".yaml")

// 4. 加载到 Manager
manager.AddTranslation(locale, key, value)
```

整个过程在内存中完成，不涉及文件系统操作。只加载 CLI 需要的 `cli.yaml` 文件，不加载其他翻译文件（如 `space.yaml`、`repo.yaml` 等），减少内存占用和加载时间。

## 注意事项

- 语言检测只在程序启动时执行一次
- 支持的语言环境格式：`zh-CN`, `zh_CN`, `zh_CN.UTF-8` 等
- 如果检测到的语言不在支持列表中，会自动回退到英文
- 所有 CLI 相关的翻译键都以 `cli.` 开头
- 修改翻译后必须运行 `task i18n:generate` 重新生成常量
- 翻译文件会被编译到二进制文件中，无需单独分发
- 二进制文件大小会增加（每个语言约几 KB）

## 常见问题

### Q: 如何查看二进制文件中包含的翻译？

A: 翻译已嵌入二进制文件，可以通过以下方式验证：

```bash
# 查看二进制文件大小
ls -lh gitfox

# 使用 strings 命令查看嵌入的文本（包含翻译）
strings gitfox | grep "启动服务器"
```

### Q: 如何添加新语言？

A: 
1. 在 `resources/locales/` 下创建新语言目录（如 `ja-JP/`）
2. 添加 `cli.yaml` 文件并翻译
3. 运行 `task i18n:generate` 生成常量
4. 在 `pkg/i18n/cli/i18n_key.go` 的 `loadTranslationsFromEmbedFS()` 函数中的 `locales` 数组添加新语言：
   ```go
   locales := []string{"en-US", "zh-CN", "ja-JP"}
   ```
5. 重新编译

### Q: 性能如何？

A: 非常快！翻译直接从内存加载，无文件 I/O：
- 首次加载：< 1ms（从 embedded FS 读取并解析）
- 后续查询：< 0.1μs（直接从内存 map 查询）
- 无临时文件创建和清理开销
