# Git 包

GitFox 社区版的 Git 操作核心包，提供完整的 Git 仓库管理和操作功能，支持分支、提交、合并、差异比较等全面的 Git 功能。

## 🚀 特性

- **完整的 Git 操作**: 支持仓库创建、分支管理、提交操作、标签管理等
- **高性能**: 基于原生 Git 命令，提供高效的操作性能
- **流式处理**: 支持大文件和大量数据的流式处理
- **缓存优化**: 内置提交缓存，提升查询性能
- **Hook 支持**: 完整的 Git Hook 机制支持
- **安全性**: 内置安全检查和验证机制
- **并发安全**: 支持并发操作和共享仓库访问

## 📦 核心模块

### 1. Repository Management (仓库管理)
- 创建和删除仓库
- 仓库大小计算
- 仓库同步和镜像
- 仓库存储管理

### 2. Branch Operations (分支操作)
- 创建、删除、列出分支
- 分支标签管理
- 默认分支设置
- 分支保护和验证

### 3. Commit Operations (提交操作)
- 文件提交和修改
- 提交历史查询
- 提交统计和分析
- 提交差异比较

### 4. File Operations (文件操作)
- 文件树浏览
- 文件内容获取
- 文件差异比较
- 文件搜索和匹配

### 5. Merge & Diff (合并与差异)
- 分支合并操作
- 差异比较和统计
- 冲突检测和解决
- Patch 应用

### 6. Advanced Features (高级功能)
- Git Blame 分析
- 子模块支持
- 归档和打包
- 安全扫描

## 🔧 核心接口

### Git Interface
```go
type Interface interface {
    // 仓库管理
    CreateRepository(ctx context.Context, params *CreateRepositoryParams) (*CreateRepositoryOutput, error)
    DeleteRepository(ctx context.Context, params *DeleteRepositoryParams) error
    GetRepositorySize(ctx context.Context, params *GetRepositorySizeParams) (*GetRepositorySizeOutput, error)
    
    // 分支操作
    CreateBranch(ctx context.Context, params *CreateBranchParams) (*CreateBranchOutput, error)
    DeleteBranch(ctx context.Context, params *DeleteBranchParams) error
    ListBranches(ctx context.Context, params *ListBranchesParams) (*ListBranchesOutput, error)
    GetBranch(ctx context.Context, params *GetBranchParams) (*GetBranchOutput, error)
    
    // 提交操作
    CommitFiles(ctx context.Context, params *CommitFilesParams) (CommitFilesResponse, error)
    GetCommit(ctx context.Context, params *GetCommitParams) (*GetCommitOutput, error)
    ListCommits(ctx context.Context, params *ListCommitsParams) (*ListCommitsOutput, error)
    
    // 文件操作
    GetTreeNode(ctx context.Context, params *GetTreeNodeParams) (*GetTreeNodeOutput, error)
    ListTreeNodes(ctx context.Context, params *ListTreeNodeParams) (*ListTreeNodeOutput, error)
    GetBlob(ctx context.Context, params *GetBlobParams) (*GetBlobOutput, error)
    
    // 差异比较
    Diff(ctx context.Context, in *DiffParams, files ...api.FileDiffRequest) (<-chan *FileDiff, <-chan error)
    RawDiff(ctx context.Context, w io.Writer, in *DiffParams, files ...api.FileDiffRequest) error
    
    // 合并操作
    Merge(ctx context.Context, in *MergeParams) (MergeOutput, error)
    
    // 其他功能
    Blame(ctx context.Context, params *BlameParams) (<-chan *BlamePart, <-chan error)
    Archive(ctx context.Context, params ArchiveParams, w io.Writer) error
}
```

### Service 结构
```go
type Service struct {
    reposRoot         string
    tmpDir            string
    git               *api.Git
    hookClientFactory hook.ClientFactory
    store             storage.Store
    gitHookPath       string
    reposGraveyard    string
}
```

## 📖 使用示例

### 创建 Git 服务

```go
package main

import (
    "context"
    
    "github.com/easysoft/gitfox-community/pkg/git"
    "github.com/easysoft/gitfox-community/pkg/git/api"
    "github.com/easysoft/gitfox-community/pkg/git/hook"
    "github.com/easysoft/gitfox-community/pkg/git/storage"
    "github.com/easysoft/gitfox-community/pkg/git/types"
)

func main() {
    // 配置
    config := types.Config{
        Root:  "/path/to/repos",
        Trace: false,
    }
    
    // 创建依赖
    gitAPI, _ := api.New(config, lastCommitCache, hookFactory)
    hookFactory := hook.NewClientFactory()
    store := storage.NewLocal()
    
    // 创建 Git 服务
    gitService, err := git.New(config, gitAPI, hookFactory, store)
    if err != nil {
        panic(err)
    }
}
```

### 仓库操作示例

```go
// 创建仓库
func createRepository(ctx context.Context, gitService git.Interface) error {
    params := &git.CreateRepositoryParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
        DefaultBranch: "main",
    }
    
    output, err := gitService.CreateRepository(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("仓库创建成功: %s\n", output.UID)
    return nil
}

// 删除仓库
func deleteRepository(ctx context.Context, gitService git.Interface) error {
    params := &git.DeleteRepositoryParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
    }
    
    return gitService.DeleteRepository(ctx, params)
}
```

### 分支操作示例

```go
// 创建分支
func createBranch(ctx context.Context, gitService git.Interface) error {
    params := &git.CreateBranchParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
        Name:   "feature/new-feature",
        Target: "main",
    }
    
    output, err := gitService.CreateBranch(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("分支创建成功: %s\n", output.Name)
    return nil
}

// 列出分支
func listBranches(ctx context.Context, gitService git.Interface) error {
    params := &git.ListBranchesParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        IncludeCommit: true,
        Sort:          git.BranchSortOptionName,
        Order:         git.SortOrderAsc,
    }
    
    output, err := gitService.ListBranches(ctx, params)
    if err != nil {
        return err
    }
    
    for _, branch := range output.Branches {
        fmt.Printf("分支: %s, 提交: %s\n", branch.Name, branch.SHA)
    }
    return nil
}

// 删除分支
func deleteBranch(ctx context.Context, gitService git.Interface) error {
    params := &git.DeleteBranchParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
        Name: "feature/old-feature",
    }
    
    return gitService.DeleteBranch(ctx, params)
}
```

### 文件操作示例

```go
// 提交文件
func commitFiles(ctx context.Context, gitService git.Interface) error {
    params := &git.CommitFilesParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
        Message: "添加新功能",
        Branch:  "main",
        Actions: []git.CommitFileAction{
            {
                Action:  git.CreateAction,
                Path:    "src/new_file.go",
                Payload: []byte("package main\n\nfunc main() {\n\tprintln(\"Hello World\")\n}"),
            },
            {
                Action: git.UpdateAction,
                Path:   "README.md",
                Payload: []byte("# 项目标题\n\n更新的描述"),
            },
        },
    }
    
    response, err := gitService.CommitFiles(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("提交成功: %s\n", response.CommitID)
    return nil
}

// 获取文件内容
func getFileContent(ctx context.Context, gitService git.Interface) error {
    params := &git.GetBlobParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        GitREF: "main",
        Path:   "src/main.go",
    }
    
    output, err := gitService.GetBlob(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("文件大小: %d bytes\n", output.Content.Size)
    fmt.Printf("文件内容: %s\n", string(output.Content.Data))
    return nil
}

// 浏览文件树
func listTreeNodes(ctx context.Context, gitService git.Interface) error {
    params := &git.ListTreeNodeParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        GitREF:            "main",
        Path:              "",
        IncludeLatestCommit: true,
    }
    
    output, err := gitService.ListTreeNodes(ctx, params)
    if err != nil {
        return err
    }
    
    for _, node := range output.Nodes {
        fmt.Printf("%s %s %s\n", node.Type, node.Name, node.Path)
        if node.LatestCommit != nil {
            fmt.Printf("  最新提交: %s\n", node.LatestCommit.Title)
        }
    }
    return nil
}
```

### 提交历史示例

```go
// 获取提交历史
func listCommits(ctx context.Context, gitService git.Interface) error {
    params := &git.ListCommitsParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        GitREF: "main",
        After:  "",
        Limit:  10,
    }
    
    output, err := gitService.ListCommits(ctx, params)
    if err != nil {
        return err
    }
    
    for _, commit := range output.Commits {
        fmt.Printf("提交: %s\n", commit.SHA)
        fmt.Printf("作者: %s <%s>\n", commit.Author.Identity.Name, commit.Author.Identity.Email)
        fmt.Printf("时间: %s\n", commit.Author.When.Format("2006-01-02 15:04:05"))
        fmt.Printf("消息: %s\n\n", commit.Title)
    }
    return nil
}

// 获取单个提交
func getCommit(ctx context.Context, gitService git.Interface) error {
    params := &git.GetCommitParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        SHA: "abc123def456",
    }
    
    output, err := gitService.GetCommit(ctx, params)
    if err != nil {
        return err
    }
    
    commit := output.Commit
    fmt.Printf("提交 SHA: %s\n", commit.SHA)
    fmt.Printf("标题: %s\n", commit.Title)
    fmt.Printf("消息: %s\n", commit.Message)
    fmt.Printf("作者: %s <%s>\n", commit.Author.Identity.Name, commit.Author.Identity.Email)
    fmt.Printf("提交者: %s <%s>\n", commit.Committer.Identity.Name, commit.Committer.Identity.Email)
    
    return nil
}
```

### 差异比较示例

```go
// 比较差异
func compareDiff(ctx context.Context, gitService git.Interface) error {
    params := &git.DiffParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        BaseRef:   "main",
        HeadRef:   "feature/new-feature",
        MergeBase: true,
    }
    
    // 获取差异文件名
    fileNames, err := gitService.DiffFileNames(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("修改的文件数量: %d\n", len(fileNames.Files))
    for _, file := range fileNames.Files {
        fmt.Printf("文件: %s (状态: %s)\n", file.Path, file.Status)
    }
    
    // 获取差异统计
    stats, err := gitService.DiffStats(ctx, params)
    if err != nil {
        return err
    }
    
    fmt.Printf("总计: +%d -%d\n", stats.Additions, stats.Deletions)
    
    return nil
}

// 流式获取差异
func streamDiff(ctx context.Context, gitService git.Interface) error {
    params := &git.DiffParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        BaseRef: "main",
        HeadRef: "feature/new-feature",
    }
    
    diffChan, errChan := gitService.Diff(ctx, params)
    
    for {
        select {
        case diff, ok := <-diffChan:
            if !ok {
                return nil
            }
            fmt.Printf("文件差异: %s\n", diff.Path)
            for _, hunk := range diff.Hunks {
                fmt.Printf("  Hunk: %s\n", hunk.Header)
            }
            
        case err := <-errChan:
            if err != nil {
                return err
            }
        }
    }
}
```

### 合并操作示例

```go
// 合并分支
func mergeBranch(ctx context.Context, gitService git.Interface) error {
    params := &git.MergeParams{
        WriteParams: git.WriteParams{
            RepoUID: "repo-123",
            Actor: git.Identity{
                Name:  "Developer",
                Email: "dev@example.com",
            },
        },
        BaseBranch:   "main",
        HeadRepoUID:  "repo-123",
        HeadBranch:   "feature/new-feature",
        CommitTitle:  "合并新功能分支",
        CommitMessage: "将新功能分支合并到主分支",
        Method:       git.MergeMethodMerge,
    }
    
    output, err := gitService.Merge(ctx, params)
    if err != nil {
        return err
    }
    
    if output.ConflictFiles != nil && len(output.ConflictFiles) > 0 {
        fmt.Printf("存在冲突文件: %v\n", output.ConflictFiles)
        return fmt.Errorf("合并存在冲突")
    }
    
    fmt.Printf("合并成功: %s\n", output.SHA)
    return nil
}
```

### Blame 分析示例

```go
// 文件 Blame 分析
func blameFile(ctx context.Context, gitService git.Interface) error {
    params := &git.BlameParams{
        ReadParams: git.ReadParams{
            RepoUID: "repo-123",
        },
        Path:   "src/main.go",
        GitREF: "main",
    }
    
    blameChan, errChan := gitService.Blame(ctx, params)
    
    for {
        select {
        case blame, ok := <-blameChan:
            if !ok {
                return nil
            }
            
            fmt.Printf("行 %d-%d: 提交 %s\n", 
                blame.StartingLine, blame.EndingLine, blame.Commit.SHA)
            fmt.Printf("  作者: %s <%s>\n", 
                blame.Commit.Author.Identity.Name, blame.Commit.Author.Identity.Email)
            fmt.Printf("  时间: %s\n", 
                blame.Commit.Author.When.Format("2006-01-02 15:04:05"))
            
            for _, line := range blame.Lines {
                fmt.Printf("  %s\n", line.Text)
            }
            
        case err := <-errChan:
            if err != nil {
                return err
            }
        }
    }
}
```

## 🔧 配置选项

### Git 配置
```go
type Config struct {
    Root  string // Git 仓库根目录
    Trace bool   // 是否启用 Git 命令跟踪
}
```

### 身份配置
```go
type Identity struct {
    Name  string // 用户名
    Email string // 邮箱地址
}
```

### 读写参数
```go
type ReadParams struct {
    RepoUID string // 仓库唯一标识
}

type WriteParams struct {
    ReadParams
    Actor Identity // 操作者身份
}
```

## 📊 性能特点

| 功能模块 | 性能特点 | 适用场景 | 注意事项 |
|----------|----------|----------|----------|
| 仓库操作 | 高效 | 仓库管理 | 注意磁盘空间 |
| 分支操作 | 极快 | 分支管理 | 轻量级操作 |
| 提交操作 | 中等 | 文件提交 | 大文件影响性能 |
| 差异比较 | 快速 | 代码审查 | 大差异集影响性能 |
| 历史查询 | 快速 | 日志查看 | 缓存优化 |
| 文件操作 | 快速 | 文件浏览 | 大文件流式处理 |

## 🛠️ 最佳实践

### 1. 错误处理
```go
output, err := gitService.CreateBranch(ctx, params)
if err != nil {
    if errors.IsConflict(err) {
        // 处理冲突错误
        return handleConflict(err)
    }
    // 处理其他错误
    return err
}
```

### 2. 上下文管理
```go
// 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := gitService.ListCommits(ctx, params)
```

### 3. 流式处理大数据
```go
// 使用流式接口处理大量差异数据
diffChan, errChan := gitService.Diff(ctx, params)
// 逐个处理，避免内存溢出
```

### 4. 缓存利用
```go
// 利用内置缓存提升性能
// 重复的提交查询会自动使用缓存
```

### 5. 并发安全
```go
// 多个 goroutine 可以安全地并发访问同一个 gitService 实例
go func() {
    gitService.GetCommit(ctx, params1)
}()
go func() {
    gitService.ListBranches(ctx, params2)
}()
```

## 🧪 测试

运行测试:
```bash
go test ./pkg/git -v
```

运行特定模块测试:
```bash
go test ./pkg/git/api -v
go test ./pkg/git/parser -v
```

运行基准测试:
```bash
go test ./pkg/git -bench=. -benchmem
```

## 📝 注意事项

1. **磁盘空间**: Git 操作需要足够的磁盘空间，特别是大仓库
2. **并发限制**: 虽然支持并发，但过多并发可能影响性能
3. **内存使用**: 大文件操作时注意内存使用，优先使用流式接口
4. **权限管理**: 确保进程有足够权限访问 Git 仓库目录
5. **Hook 配置**: 正确配置 Git Hook 以确保完整功能
6. **错误处理**: 妥善处理各种 Git 操作错误，特别是冲突和权限错误

## 🔗 相关包

- `pkg/cache`: 提供缓存支持
- `pkg/errors`: 错误处理
- `pkg/logger`: 日志记录

## 📄 许可证

本包基于 Apache License 2.0 许可证开源。