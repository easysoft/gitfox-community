# Storage Package

Storage 包提供了一个可插拔的存储驱动抽象层，用于管理文件的读写、移动、删除等操作。该设计参考了 Docker Distribution 的存储驱动架构，支持多种后端存储实现。

## 目录结构

```
pkg/storage/
├── storage.go              # 驱动创建入口
├── content_storage.go      # ContentStorage 接口及实现
├── content_storage_option.go # ContentStorage 配置选项
├── file_reader.go          # 带缓冲的文件读取器
├── config.go               # 配置类型定义
├── wire.go                 # 依赖注入 (Wire)
└── driver/
    ├── storagedriver.go    # StorageDriver 核心接口
    ├── fileinfo.go         # FileInfo 接口
    ├── walk.go             # 目录遍历实现
    ├── factory/
    │   └── factory.go      # 驱动工厂 (注册/创建)
    ├── base/
    │   ├── base.go         # 基础包装器 (路径验证)
    │   └── regulator.go    # 并发调节器
    ├── filesystem/
    │   └── driver.go       # 本地文件系统驱动
    └── inmemory/
        ├── driver.go       # 内存驱动
        └── mfs.go          # 内存文件系统实现
```

## 核心接口

### StorageDriver

`StorageDriver` 是最底层的存储驱动接口，定义了文件系统式的键值存储操作：

```go
type StorageDriver interface {
    Name() string
    GetContent(ctx context.Context, path string) ([]byte, error)
    PutContent(ctx context.Context, path string, content []byte) error
    Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)
    Writer(ctx context.Context, path string, append bool) (FileWriter, error)
    Stat(ctx context.Context, path string) (FileInfo, error)
    List(ctx context.Context, path string) ([]string, error)
    Move(ctx context.Context, sourcePath string, destPath string) error
    Delete(ctx context.Context, path string) error
    RedirectURL(r *http.Request, path string) (string, error)
    Walk(ctx context.Context, path string, f WalkFn, options ...func(*WalkOptions)) error
}
```

### ContentStorage

`ContentStorage` 是更高层的内容存储接口，封装了 `StorageDriver`，提供了更便捷的文件操作方法：

```go
type ContentStorage interface {
    Get(ctx context.Context, path string) ([]byte, error)
    Open(ctx context.Context, path string) (io.ReadCloser, error)
    Put(ctx context.Context, path string, contents []byte) error
    Save(ctx context.Context, path string, r io.Reader, size int64) error
    GetWriter(ctx context.Context, path string, append bool) (FileWriter, error)
    Stat(ctx context.Context, path string) (FileInfo, error)
    Delete(ctx context.Context, path string) error
    List(ctx context.Context, path string) ([]string, error)
    PrefixDir() string
}
```

## 架构设计

### 分层架构

```
+---------------------------+
|     ContentStorage        |  <- 应用层接口
+---------------------------+
            |
+---------------------------+
|    Base (Wrapper)         |  <- 路径验证、错误处理
+---------------------------+
            |
+---------------------------+
|    Regulator              |  <- 并发控制
+---------------------------+
            |
+---------------------------+
|    StorageDriver          |  <- 具体实现
|  (filesystem/inmemory)    |
+---------------------------+
```

### 工厂模式

驱动通过工厂模式注册和创建：

```go
// 驱动在 init() 中自动注册
func init() {
    factory.Register("filesystem", &filesystemDriverFactory{})
}

// 创建驱动
driver, err := storage.NewDriver(ctx, "filesystem", map[string]interface{}{
    "rootdirectory": "/var/lib/gitfox",
    "maxthreads":    100,
})
```

### 并发控制

`Regulator` 组件通过信号量机制限制对底层驱动的并发访问，防止创建过多的 OS 线程：

```go
type regulator struct {
    StorageDriver
    *sync.Cond
    available uint64  // 可用的并发槽位
}
```

## 驱动实现

### Filesystem Driver

本地文件系统驱动，适用于生产环境：

- **默认根目录**: `/var/lib/gitfox`
- **默认最大线程数**: 100
- **最小线程数**: 25

配置参数：
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `rootdirectory` | string | `/var/lib/gitfox` | 存储根目录 |
| `maxthreads` | uint64 | 100 | 最大并发数 |

### InMemory Driver

内存驱动，主要用于测试场景：

- 基于内存中的目录树结构实现
- 使用读写锁保证并发安全
- 不持久化，进程结束后数据丢失

## 使用示例

### 创建 ContentStorage

```go
import (
    "context"
    "github.com/easysoft/gitfox-community/pkg/storage"
)

func main() {
    ctx := context.Background()

    // 创建 filesystem 驱动
    driver, err := storage.NewDriver(ctx, "filesystem", map[string]interface{}{
        "rootdirectory": "/data/storage",
    })
    if err != nil {
        panic(err)
    }

    // 创建 ContentStorage，设置路径前缀
    cs, err := storage.NewCommonContentStore(ctx, driver,
        storage.WithPrefix("/uploads"))
    if err != nil {
        panic(err)
    }

    // 写入文件
    err = cs.Put(ctx, "/example.txt", []byte("Hello, World!"))

    // 读取文件
    content, err := cs.Get(ctx, "/example.txt")

    // 流式写入
    err = cs.Save(ctx, "/large-file.bin", reader, fileSize)

    // 流式读取
    rc, err := cs.Open(ctx, "/large-file.bin")
    defer rc.Close()

    // 列出目录
    files, err := cs.List(ctx, "/")

    // 删除文件
    err = cs.Delete(ctx, "/example.txt")
}
```

### 使用内存驱动进行测试

```go
import (
    "github.com/easysoft/gitfox-community/pkg/storage/driver/inmemory"
)

func TestStorage(t *testing.T) {
    driver := inmemory.New()
    cs, _ := storage.NewCommonContentStore(context.Background(), driver)

    // 测试逻辑...
}
```

## 依赖注入 (Wire)

Storage 包支持通过 Google Wire 进行依赖注入，便于与项目其他模块集成。

### 配置

在 `config.yaml` 中添加存储配置：

```yaml
storage:
  driver: filesystem           # 存储驱动: filesystem, inmemory
  rootDirectory: /data/gitfox/storage  # 存储根目录
  maxThreads: 100              # 最大并发线程数
  pathPrefix: /uploads         # 路径前缀（可选）
```

或通过环境变量配置：

```bash
export GITFOX_STORAGE_ROOT_DIR=/data/gitfox/storage
```

### 在服务中使用

1. 在 Wire 配置中引入 `StorageWireSet`：

```go
// cmd/cli/operations/server/wire.go
func NewApp(mode string) (*App, func(), error) {
    panic(wire.Build(
        // ... 其他 WireSet
        provider.StorageWireSet,
        wire.Struct(new(App), "*")),
    )
}
```

2. 在 App 结构体中声明依赖：

```go
type App struct {
    // ... 其他字段
    ContentStorage storage.ContentStorage
}
```

3. 在业务代码中直接使用注入的 `ContentStorage`：

```go
type MyService struct {
    storage storage.ContentStorage
}

func NewMyService(cs storage.ContentStorage) *MyService {
    return &MyService{storage: cs}
}

func (s *MyService) SaveFile(ctx context.Context, path string, data []byte) error {
    return s.storage.Put(ctx, path, data)
}
```

### WireSet 说明

`storage.WireSet` 提供以下 Provider：

| Provider | 输出类型 | 说明 |
|----------|----------|------|
| `ProvideStorageDriver` | `driver.StorageDriver` | 根据配置创建存储驱动 |
| `ProvideContentStorage` | `ContentStorage` | 创建内容存储服务 |

## 路径规则

所有路径必须符合以下正则表达式：

```
^(/[A-Za-z0-9._-]+)+$
```

- 必须以 `/` 开头
- 每个路径组件只能包含字母、数字、点、下划线和连字符
- 不支持相对路径

## 错误类型

| 错误类型 | 说明 |
|----------|------|
| `PathNotFoundError` | 路径不存在 |
| `InvalidPathError` | 路径格式无效 |
| `InvalidOffsetError` | 读取偏移量无效 |
| `ErrUnsupportedMethod` | 驱动不支持的方法 |

## 测试

运行所有测试：

```bash
go test -v ./pkg/storage/...
```

运行特定驱动的测试：

```bash
# Filesystem 驱动测试
go test -v ./pkg/storage/driver/filesystem/...

# InMemory 驱动测试
go test -v ./pkg/storage/driver/inmemory/...
```

## 扩展新驱动

要添加新的存储驱动：

1. 实现 `StorageDriver` 接口
2. 在 `init()` 函数中调用 `factory.Register()` 注册驱动
3. 在 `storage.go` 中匿名导入新驱动包

```go
package mydriver

import (
    "github.com/easysoft/gitfox-community/pkg/storage/driver"
    "github.com/easysoft/gitfox-community/pkg/storage/driver/factory"
)

func init() {
    factory.Register("mydriver", &myDriverFactory{})
}

type myDriverFactory struct{}

func (f *myDriverFactory) Create(ctx context.Context, parameters map[string]interface{}) (driver.StorageDriver, error) {
    // 创建并返回驱动实例
}
```
