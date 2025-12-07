# Logger 包使用文档

基于 Uber Zap 封装的统一日志接口。

## 特性

- ✅ 结构化日志（支持 JSON 和 Console 格式）
- ✅ 多级别日志（Debug, Info, Warn, Error, Fatal）
- ✅ 灵活配置（输出路径、格式、级别）
- ✅ Gin 中间件集成
- ✅ 彩色控制台输出
- ✅ 自动记录调用位置
- ✅ 性能优化（零内存分配）
- ✅ 线程安全（使用读写锁保护）
- ✅ 支持多输出目标（同时输出到控制台和文件）

## 快速开始

### 1. 初始化 Logger

在 `main.go` 中初始化：

```go
import "neverbroke/common/logger"

func main() {
    // 初始化日志 - 同时输出到控制台和文件
    if err := logger.Init(logger.Config{
        Level: "info",      // debug, info, warn, error
        OutputPaths: []string{
            "stdout",          // 输出到控制台
            "./logs/app.log",  // 同时输出到文件
        },
        Format: "console",   // console 或 json
    }); err != nil {
        panic(err)
    }
    defer logger.Sync()
    
    // ... 其他代码
}
```

### 2. 基本使用

```go
import (
    "neverbroke/common/logger"
    "go.uber.org/zap"
)

// 简单日志
logger.Info("Server started")
logger.Debug("Debug message")
logger.Warn("Warning message")
logger.Error("Error occurred")

// 带字段的日志
logger.Info("User logged in",
    zap.String("userID", "123"),
    zap.String("ip", "192.168.1.1"),
)

logger.Error("Database connection failed",
    zap.Error(err),
    zap.String("host", "localhost"),
    zap.Int("port", 5432),
)
```

### 3. Gin 中间件

在路由设置中使用：

```go
import (
    "neverbroke/common/logger"
    "github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
    r := gin.New()
    
    // 使用日志中间件
    r.Use(logger.GinLogger())    // 记录 HTTP 请求
    r.Use(logger.GinRecovery())  // Panic 恢复
    
    // ... 注册路由
    
    return r
}
```

## 配置说明

### Logger.Config

```go
type Config struct {
    Level       string   // 日志级别：debug, info, warn, error
    OutputPaths []string // 输出路径列表：可包含 stdout, stderr, 或文件路径，支持多个同时输出
    Format      string   // 输出格式：console 或 json
}
```

**支持同时输出到多个目标**：
- 可以同时输出到控制台和文件
- 控制台输出会自动使用彩色格式
- 文件输出推荐使用 JSON 格式便于日志分析

### 日志级别

- `debug`: 调试信息，开发环境使用
- `info`: 一般信息，生产环境推荐
- `warn`: 警告信息
- `error`: 错误信息

### 输出格式

**Console 格式**（开发环境推荐）：
```
2024-12-07T10:30:45.123+0800    INFO    home/handler.go:25    Home overview fetched successfully
```

**JSON 格式**（生产环境推荐）：
```json
{
  "time": "2024-12-07T10:30:45.123+0800",
  "level": "INFO",
  "caller": "home/handler.go:25",
  "msg": "Home overview fetched successfully"
}
```

## API 参考

### 全局函数

```go
// 初始化全局 logger
func Init(cfg Config) error

// 获取全局 logger
func GetLogger() *Logger

// 日志记录
func Debug(msg string, fields ...zap.Field)
func Info(msg string, fields ...zap.Field)
func Warn(msg string, fields ...zap.Field)
func Error(msg string, fields ...zap.Field)
func Fatal(msg string, fields ...zap.Field)

// 创建带额外字段的 logger
func With(fields ...zap.Field) *Logger

// 刷新日志缓冲区
func Sync() error

// 设置全局 logger（高级用法）
func SetLogger(logger *Logger)

// 重置全局 logger（主要用于测试）
func ResetLogger()
```

### Zap 常用字段

```go
import "go.uber.org/zap"

// 字符串
zap.String("key", "value")

// 数字
zap.Int("count", 10)
zap.Int64("id", 123456)
zap.Float64("price", 99.99)

// 布尔值
zap.Bool("success", true)

// 错误
zap.Error(err)

// 时间
zap.Time("timestamp", time.Now())
zap.Duration("latency", time.Second)

// 任意类型（会使用反射，性能较低）
zap.Any("data", someStruct)
```

## 使用示例

### Handler 中使用

```go
package home

import (
    "neverbroke/pkg/logger"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func (h *Handler) GetOverview(c *gin.Context) {
    logger.Debug("Fetching home overview")
    
    // 业务逻辑
    data, err := h.service.GetData()
    if err != nil {
        logger.Error("Failed to get data",
            zap.Error(err),
            zap.String("handler", "GetOverview"),
        )
        return
    }
    
    logger.Info("Home overview fetched successfully",
        zap.Int("itemCount", len(data)),
    )
}
```

### 带上下文的 Logger

```go
// 创建带固定字段的 logger
userLogger := logger.With(
    zap.String("userID", "123"),
    zap.String("requestID", "abc-def"),
)

// 后续日志都会自动带上这些字段
userLogger.Info("User action performed")
userLogger.Error("User action failed", zap.Error(err))
```

### 只输出到文件

```go
logger.Init(logger.Config{
    Level:       "info",
    OutputPaths: []string{"./logs/app.log"},  // 只输出到文件
    Format:      "json",                      // JSON 格式便于日志分析
})
```

### 同时输出到控制台和文件（推荐）

```go
logger.Init(logger.Config{
    Level: "info",
    OutputPaths: []string{
        "stdout",          // 控制台输出（彩色）
        "./logs/app.log",  // 文件输出
    },
    Format: "console",  // 或 "json"
})
```

### 输出到多个文件

```go
logger.Init(logger.Config{
    Level: "info",
    OutputPaths: []string{
        "stdout",
        "./logs/app.log",
        "./logs/error.log",  // 可以输出到多个文件
    },
    Format: "json",
})
```

### 不同环境配置

```go
var logConfig logger.Config

if os.Getenv("ENV") == "production" {
    // 生产环境：同时输出到控制台和文件
    logConfig = logger.Config{
        Level: "info",
        OutputPaths: []string{
            "stdout",
            "./logs/app.log",
        },
        Format: "json",
    }
} else {
    // 开发环境：只输出到控制台
    logConfig = logger.Config{
        Level:       "debug",
        OutputPaths: []string{"stdout"},
        Format:      "console",
    }
}

logger.Init(logConfig)
```

## 线程安全

本 logger 包是**完全线程安全**的：

- ✅ 使用 `sync.RWMutex` 保护全局 logger
- ✅ 支持并发读取（多个 goroutine 同时记录日志）
- ✅ 初始化时使用双重检查锁定（Double-Checked Locking）
- ✅ Zap 本身也是线程安全的

**并发使用示例**：
```go
// 多个 goroutine 可以安全地同时记录日志
for i := 0; i < 100; i++ {
    go func(id int) {
        logger.Info("Concurrent log", zap.Int("goroutine", id))
    }(i)
}
```

## 性能说明

Zap 是高性能日志库：

- 零内存分配（使用预分配的字段）
- 比标准库 `log` 快 10x+
- 比其他结构化日志库快 4-10x
- 适合高并发场景
- 读写锁开销极小（读操作无竞争）

**推荐做法**：
- ✅ 使用 `zap.String()` 等类型化字段
- ❌ 避免使用 `zap.Any()` 和字符串拼接
- ✅ 生产环境使用 `info` 级别
- ✅ 使用 JSON 格式便于日志分析
- ✅ 在高并发场景下放心使用

## 最佳实践

1. **合理使用日志级别**
   - Debug: 详细的调试信息
   - Info: 重要的业务流程
   - Warn: 需要注意但不影响运行
   - Error: 错误但可恢复
   - Fatal: 致命错误，程序退出

2. **添加有用的上下文**
   ```go
   logger.Info("Order created",
       zap.String("orderID", order.ID),
       zap.String("userID", user.ID),
       zap.Float64("amount", order.Amount),
   )
   ```

3. **不要记录敏感信息**
   - ❌ 密码、Token、信用卡号
   - ✅ 用户 ID、订单号、操作类型

4. **错误日志包含足够信息**
   ```go
   logger.Error("Failed to process payment",
       zap.Error(err),
       zap.String("orderID", orderID),
       zap.String("paymentMethod", method),
   )
   ```

5. **在程序退出前刷新日志**
   ```go
   func main() {
       logger.Init(config)
       defer logger.Sync()  // 确保所有日志都写入
       
       // ... 程序逻辑
   }
   ```

## 高级用法

### 自定义 Logger 实例

```go
// 创建自定义配置的 logger
customLogger, err := logger.NewLogger(logger.Config{
    Level:       "debug",
    OutputPaths: []string{"./logs/custom.log"},
    Format:      "json",
})
if err != nil {
    panic(err)
}

// 使用自定义 logger
customLogger.Info("Custom logger message")

// 或者设置为全局 logger
logger.SetLogger(customLogger)
```

### 测试中重置 Logger

```go
func TestSomething(t *testing.T) {
    // 测试前重置
    logger.ResetLogger()
    
    // 使用测试配置
    logger.Init(logger.Config{
        Level:       "debug",
        OutputPaths: []string{"stdout"},
        Format:      "console",
    })
    
    // ... 测试逻辑
    
    // 测试后清理
    defer logger.ResetLogger()
}
```

### 并发安全性验证

```go
// 压力测试：1000 个并发 goroutine 同时写日志
func BenchmarkConcurrentLogging(b *testing.B) {
    logger.Init(logger.Config{
        Level:       "info",
        OutputPaths: []string{"./logs/bench.log"},
        Format:      "json",
    })
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            logger.Info("Concurrent log",
                zap.Int("iteration", b.N),
                zap.String("data", "test"),
            )
        }
    })
}
```

## 许可证

MIT License

