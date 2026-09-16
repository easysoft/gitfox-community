# GitFox Errors

GitFox Errors 是一个基于 `github.com/pkg/errors` 的增强错误处理包，提供了类型化错误、错误上下文和便捷的错误处理方法。

## 特性

- 🏷️ **类型化错误**：支持预定义的错误类型，便于错误分类和处理
- 📝 **错误上下文**：可以为错误添加字段和消息上下文
- 🔗 **错误包装**：支持错误包装和链式调用
- 🎯 **错误判断**：提供便捷的错误类型判断方法
- 🚀 **高性能**：基于成熟的 pkg/errors 库，性能优异
- 🛡️ **向下兼容**：完全兼容标准库的 error 接口


## 快速开始

### 1. 基本错误创建

```go
package main

import (
    "fmt"
    pkg_errors "github.com/easysoft/gitfox-community/pkg/errors"
)

func main() {
    // 创建无类型错误
    err1 := pkg_errors.New("something went wrong")
    fmt.Println(err1) // something went wrong
    
    // 创建格式化错误
    err2 := pkg_errors.Newf("user %s not found", "john")
    fmt.Println(err2) // user john not found
    
    // 创建类型化错误
    err3 := pkg_errors.BadRequest.New("invalid parameter")
    fmt.Println(pkg_errors.GetType(err3)) // 1 (BadRequest)
}
```

### 2. 错误类型

预定义的错误类型：

```go
const (
    NoType ErrorType = iota    // 0 - 无类型错误
    BadRequest                 // 1 - 请求参数错误
    NotFound                   // 2 - 资源未找到
    DBError                    // 3 - 数据库错误
    SystemErr                  // 4 - 系统错误
)
```

### 3. 类型化错误操作

```go
// 创建类型化错误
err := pkg_errors.NotFound.New("user not found")

// 创建格式化类型化错误
err = pkg_errors.BadRequest.Newf("invalid parameter: %s", "username")

// 包装现有错误
originalErr := errors.New("connection failed")
wrappedErr := pkg_errors.DBError.Wrap(originalErr, "database operation failed")

// 格式化包装错误
wrappedErr = pkg_errors.DBError.Wrapf(originalErr, "failed to connect to %s", "mysql")
```

### 4. 错误包装和消息

```go
originalErr := errors.New("file not found")

// 包装错误
err1 := pkg_errors.Wrap(originalErr, "failed to read config")

// 格式化包装错误
err2 := pkg_errors.Wrapf(originalErr, "failed to read config from %s", "/etc/config.yaml")

// 添加消息
err3 := pkg_errors.WithMessage(originalErr, "additional context")

// 格式化添加消息
err4 := pkg_errors.WithMessagef(originalErr, "operation failed at %v", time.Now())
```

### 5. 错误上下文

```go
// 创建基础错误
err := pkg_errors.BadRequest.New("validation failed")

// 添加上下文
errWithContext := pkg_errors.AddErrorContext(err, "username", "username is required")

// 获取上下文
context := pkg_errors.GetErrorContext(errWithContext)
if context != nil {
    fmt.Println("Field:", context["field"])     // Field: username
    fmt.Println("Message:", context["message"]) // Message: username is required
}
```

### 6. 错误判断和处理

```go
err := pkg_errors.NotFound.New("user not found")

// 判断错误类型
if pkg_errors.GetType(err) == pkg_errors.NotFound {
    fmt.Println("这是一个NotFound错误")
}

// 便捷的类型判断
if pkg_errors.IsNotFound(err) {
    fmt.Println("资源未找到")
}

// 错误比较
targetErr := errors.New("target error")
wrappedErr := pkg_errors.Wrap(targetErr, "wrapped")
if pkg_errors.Is(wrappedErr, targetErr) {
    fmt.Println("错误匹配")
}

// 获取原始错误
cause := pkg_errors.Cause(wrappedErr)
fmt.Println(cause) // target error
```

## 完整示例

### Web API 错误处理

```go
package main

import (
    "database/sql"
    "fmt"
    "net/http"
    
    pkg_errors "github.com/easysoft/gitfox-community/pkg/errors"
)

type UserService struct {
    db *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, pkg_errors.BadRequest.Newf("invalid user id: %d", id)
    }
    
    user, err := s.queryUser(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, pkg_errors.NotFound.Newf("user %d not found", id)
        }
        return nil, pkg_errors.DBError.Wrap(err, "failed to query user")
    }
    
    return user, nil
}

func (s *UserService) CreateUser(user *User) error {
    if user.Email == "" {
        err := pkg_errors.BadRequest.New("validation failed")
        return pkg_errors.AddErrorContext(err, "email", "email is required")
    }
    
    if err := s.insertUser(user); err != nil {
        return pkg_errors.DBError.Wrapf(err, "failed to create user %s", user.Name)
    }
    
    return nil
}

// HTTP 处理器
func (s *UserService) HandleGetUser(w http.ResponseWriter, r *http.Request) {
    // ... 解析参数
    
    user, err := s.GetUser(userID)
    if err != nil {
        s.handleError(w, err)
        return
    }
    
    // ... 返回用户数据
}

func (s *UserService) handleError(w http.ResponseWriter, err error) {
    switch pkg_errors.GetType(err) {
    case pkg_errors.BadRequest:
        w.WriteHeader(http.StatusBadRequest)
        if context := pkg_errors.GetErrorContext(err); context != nil {
            fmt.Fprintf(w, "Bad Request: %s - %s", context["field"], context["message"])
        } else {
            fmt.Fprintf(w, "Bad Request: %s", err.Error())
        }
    case pkg_errors.NotFound:
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "Not Found: %s", err.Error())
    case pkg_errors.DBError:
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "Internal Server Error: database operation failed")
        // 记录详细错误到日志
        fmt.Printf("Database error: %+v\n", err)
    default:
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, "Internal Server Error")
        fmt.Printf("Unknown error: %+v\n", err)
    }
}
```

### 业务逻辑错误处理

```go
func ProcessOrder(orderID string) error {
    // 验证订单ID
    if orderID == "" {
        return pkg_errors.BadRequest.New("order ID is required")
    }
    
    // 查询订单
    order, err := getOrder(orderID)
    if err != nil {
        if pkg_errors.IsNotFound(err) {
            return pkg_errors.NotFound.Wrapf(err, "order %s not found", orderID)
        }
        return pkg_errors.DBError.Wrap(err, "failed to query order")
    }
    
    // 验证订单状态
    if order.Status != "pending" {
        err := pkg_errors.BadRequest.New("invalid order status")
        return pkg_errors.AddErrorContext(err, "status", 
            fmt.Sprintf("expected 'pending', got '%s'", order.Status))
    }
    
    // 处理支付
    if err := processPayment(order); err != nil {
        return pkg_errors.SystemErr.Wrapf(err, "payment processing failed for order %s", orderID)
    }
    
    return nil
}

func main() {
    if err := ProcessOrder("ORDER-123"); err != nil {
        switch pkg_errors.GetType(err) {
        case pkg_errors.BadRequest:
            fmt.Printf("请求参数错误: %s\n", err.Error())
            if context := pkg_errors.GetErrorContext(err); context != nil {
                fmt.Printf("字段: %s, 消息: %s\n", context["field"], context["message"])
            }
        case pkg_errors.NotFound:
            fmt.Printf("资源未找到: %s\n", err.Error())
        case pkg_errors.DBError:
            fmt.Printf("数据库错误: %s\n", err.Error())
        case pkg_errors.SystemErr:
            fmt.Printf("系统错误: %s\n", err.Error())
        default:
            fmt.Printf("未知错误: %s\n", err.Error())
        }
        
        // 打印错误堆栈
        fmt.Printf("错误详情: %+v\n", err)
    }
}
```

## API 参考

### 错误类型方法

```go
// 创建新错误
func (errorType ErrorType) New(msg string) error
func (errorType ErrorType) Newf(msg string, args ...interface{}) error

// 包装错误
func (errorType ErrorType) Wrap(err error, msg string) error
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error
```

### 全局函数

```go
// 创建错误
func New(msg string) error
func Newf(msg string, args ...interface{}) error

// 包装错误
func Wrap(err error, msg string) error
func Wrapf(err error, msg string, args ...interface{}) error

// 添加消息
func WithMessage(err error, msg string) error
func WithMessagef(err error, msg string, args ...interface{}) error

// 错误上下文
func AddErrorContext(err error, field, message string) error
func GetErrorContext(err error) map[string]string

// 错误判断
func GetType(err error) ErrorType
func IsNotFound(err error) bool
func Is(err, target error) bool
func Cause(err error) error
```

### 工具函数

```go
// 获取函数名（用于调试）
func GetFuncName() string
```

## 最佳实践

### 1. 错误类型选择

- **BadRequest**: 用于客户端请求参数错误
- **NotFound**: 用于资源未找到
- **DBError**: 用于数据库操作错误
- **SystemErr**: 用于系统级错误
- **NoType**: 用于通用错误

### 2. 错误包装

```go
// 好的做法：保留原始错误信息
if err := db.Query(...); err != nil {
    return pkg_errors.DBError.Wrap(err, "failed to query users")
}

// 避免：丢失原始错误信息
if err := db.Query(...); err != nil {
    return pkg_errors.DBError.New("query failed") // 丢失了原始错误
}
```

### 3. 错误上下文

```go
// 为验证错误添加具体的字段信息
if user.Email == "" {
    err := pkg_errors.BadRequest.New("validation failed")
    return pkg_errors.AddErrorContext(err, "email", "email is required")
}
```

### 4. 错误处理

```go
// 使用类型判断而不是字符串匹配
if pkg_errors.IsNotFound(err) {
    // 处理未找到错误
} else if pkg_errors.GetType(err) == pkg_errors.BadRequest {
    // 处理请求错误
}
```

## 性能考虑

- 错误创建和包装的性能开销很小
- 错误上下文使用 map 存储，适合少量字段
- 建议在性能敏感的代码路径中谨慎使用格式化错误

## 运行测试

```bash
# 运行所有测试
go test ./pkg/errors

# 运行测试并显示覆盖率
go test -cover ./pkg/errors

# 运行基准测试
go test -bench=. ./pkg/errors

# 运行特定测试
go test -run TestErrorType_New ./pkg/errors
```

## 依赖

- [github.com/pkg/errors](https://github.com/pkg/errors) - 错误处理基础库

## 许可证

本项目采用 MIT 许可证。