# Cache 缓存包

GitFox 社区版的通用缓存抽象层，提供多种缓存实现策略，支持泛型和类型安全的缓存操作。

## 🚀 特性

- **泛型支持**: 基于 Go 1.18+ 泛型，提供类型安全的缓存操作
- **多种实现**: 支持内存缓存、Redis 缓存和无缓存模式
- **TTL 支持**: 内置 TTL（Time To Live）过期机制
- **批量操作**: 支持批量获取和映射操作
- **统计信息**: 提供缓存命中率统计
- **扩展性**: 易于扩展的接口设计

## 📦 缓存实现

### 1. TTL Cache (内存缓存)
基于内存的 TTL 缓存，适用于单机部署场景。

**特点:**
- 自动过期清理
- 后台清理协程
- 无容量限制（适合短期缓存）
- 高性能读写

### 2. Redis Cache (分布式缓存)
基于 Redis 的分布式缓存，适用于集群部署场景。

**特点:**
- 分布式共享
- 持久化支持
- 可配置过期时间
- 支持自定义编解码器

### 3. No Cache (直通模式)
无缓存实现，直接调用数据源，适用于调试或特殊场景。

**特点:**
- 零缓存开销
- 直接数据访问
- 便于调试和测试

## 🔧 核心接口

### Cache 接口
```go
type Cache[K any, V any] interface {
    Stats() (int64, int64)                    // 获取统计信息 (命中数, 未命中数)
    Get(ctx context.Context, key K) (V, error) // 获取缓存值
}
```

### ExtendedCache 接口
```go
type ExtendedCache[K comparable, V Identifiable[K]] interface {
    Cache[K, V]
    Map(ctx context.Context, keys []K) (map[K]V, error) // 批量获取
}
```

### Getter 接口
```go
type Getter[K any, V any] interface {
    Find(ctx context.Context, key K) (V, error) // 数据源查找
}

type ExtendedGetter[K comparable, V Identifiable[K]] interface {
    Getter[K, V]
    FindMany(ctx context.Context, keys []K) ([]V, error) // 批量查找
}
```

## 📖 使用示例

### TTL Cache 示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

// 定义数据结构
type User struct {
    ID   int
    Name string
}

func (u User) Identifier() int {
    return u.ID
}

// 实现 Getter 接口
type UserGetter struct{}

func (g UserGetter) Find(ctx context.Context, id int) (User, error) {
    // 模拟数据库查询
    fmt.Printf("从数据库查询用户: %d\n", id)
    return User{ID: id, Name: fmt.Sprintf("User%d", id)}, nil
}

func main() {
    // 创建 TTL 缓存，过期时间 5 分钟
    userCache := cache.New[int, User](UserGetter{}, 5*time.Minute)
    defer userCache.Stop()
    
    ctx := context.Background()
    
    // 第一次获取 - 从数据源加载
    user1, err := userCache.Get(ctx, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取用户: %+v\n", user1)
    
    // 第二次获取 - 从缓存返回
    user2, err := userCache.Get(ctx, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取用户: %+v\n", user2)
    
    // 查看统计信息
    hits, misses := userCache.Stats()
    fmt.Printf("缓存统计 - 命中: %d, 未命中: %d\n", hits, misses)
}
```

### Redis Cache 示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

// JSON 编解码器
type JSONCodec[V any] struct{}

func (c JSONCodec[V]) Encode(value V) string {
    data, _ := json.Marshal(value)
    return string(data)
}

func (c JSONCodec[V]) Decode(encoded string) (V, error) {
    var value V
    err := json.Unmarshal([]byte(encoded), &value)
    return value, err
}

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建 Redis 缓存
    userCache := cache.NewRedis[int, User](
        rdb,
        UserGetter{},
        func(id int) string { return fmt.Sprintf("user:%d", id) }, // 键编码器
        JSONCodec[User]{},                                         // 编解码器
        10*time.Minute,                                           // 过期时间
        func(ctx context.Context, err error) {                    // 错误日志
            fmt.Printf("Redis 错误: %v\n", err)
        },
    )
    
    ctx := context.Background()
    user, err := userCache.Get(ctx, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取用户: %+v\n", user)
}
```

### 扩展缓存示例

```go
package main

import (
    "context"
    "time"
    
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

// 实现 ExtendedGetter 接口
type ExtendedUserGetter struct{}

func (g ExtendedUserGetter) Find(ctx context.Context, id int) (User, error) {
    return User{ID: id, Name: fmt.Sprintf("User%d", id)}, nil
}

func (g ExtendedUserGetter) FindMany(ctx context.Context, ids []int) ([]User, error) {
    users := make([]User, len(ids))
    for i, id := range ids {
        users[i] = User{ID: id, Name: fmt.Sprintf("User%d", id)}
    }
    return users, nil
}

func main() {
    // 创建扩展缓存
    userCache := cache.NewExtended[int, User](ExtendedUserGetter{}, 5*time.Minute)
    defer userCache.Stop()
    
    ctx := context.Background()
    
    // 批量获取用户
    userMap, err := userCache.Map(ctx, []int{1, 2, 3})
    if err != nil {
        panic(err)
    }
    
    for id, user := range userMap {
        fmt.Printf("用户 %d: %+v\n", id, user)
    }
}
```

### No Cache 示例

```go
package main

import (
    "context"
    "fmt"
    
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

func main() {
    // 创建无缓存实例
    userCache := cache.NewNoCache[int, User](UserGetter{})
    
    ctx := context.Background()
    
    // 每次都会调用数据源
    user1, _ := userCache.Get(ctx, 1)
    user2, _ := userCache.Get(ctx, 1) // 再次查询数据源
    
    fmt.Printf("用户1: %+v\n", user1)
    fmt.Printf("用户2: %+v\n", user2)
    
    // 统计信息始终为 0
    hits, misses := userCache.Stats()
    fmt.Printf("统计: 命中=%d, 未命中=%d\n", hits, misses)
}
```

## 📊 性能特点

| 缓存类型 | 读性能 | 写性能 | 内存使用 | 分布式 | 持久化 |
|----------|--------|--------|----------|--------|--------|
| TTL Cache | 极高 | 极高 | 中等 | ❌ | ❌ |
| Redis Cache | 高 | 高 | 低 | ✅ | ✅ |
| No Cache | 低 | 低 | 极低 | N/A | N/A |

## 🛠️ 最佳实践

### 1. 选择合适的缓存类型
- **单机部署**: 使用 TTL Cache
- **集群部署**: 使用 Redis Cache  
- **调试测试**: 使用 No Cache

### 2. 设置合理的过期时间
```go
// 根据数据更新频率设置 TTL
userCache := cache.New[int, User](getter, 10*time.Minute)  // 用户信息
configCache := cache.New[string, Config](getter, 1*time.Hour) // 配置信息
```

### 3. 监控缓存统计
```go
hits, misses := cache.Stats()
hitRate := float64(hits) / float64(hits + misses)
fmt.Printf("缓存命中率: %.2f%%\n", hitRate*100)
```

### 4. 优雅关闭
```go
defer cache.Stop() // 停止后台清理协程
```

### 5. 错误处理
```go
user, err := cache.Get(ctx, userID)
if err != nil {
    // 处理缓存获取错误
    log.Printf("缓存获取失败: %v", err)
    return err
}
```

## 🧪 测试

运行测试:
```bash
go test ./pkg/cache -v
```

运行基准测试:
```bash
go test ./pkg/cache -bench=. -benchmem
```

## 📝 注意事项

1. **TTL Cache**: 无容量限制，适合短期缓存，长期运行需要监控内存使用
2. **Redis Cache**: 需要 Redis 服务可用，注意网络延迟影响
3. **泛型约束**: 扩展缓存要求值类型实现 `Identifiable` 接口
4. **并发安全**: 所有实现都是并发安全的
5. **资源清理**: TTL Cache 需要调用 `Stop()` 方法停止后台协程

## 📄 许可证

本包基于 Apache License 2.0 许可证开源。