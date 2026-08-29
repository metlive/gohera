# gohera

基于 [Gin](https://github.com/gin-gonic/gin) 的轻量级 Web 框架，集成项目开发中常用的基础设施，开箱即用。

## 核心能力

| 模块        | 说明                                                                                                                    |
|-------------|-------------------------------------------------------------------------------------------------------------------------|
| 配置        | 基于 [Koanf](https://github.com/knadh/koanf)，自动发现配置文件，支持 TOML / YAML / JSON；可选 Nacos（HTTP/gRPC）合并与热更新 |
| 日志        | 基于 [zap](https://github.com/uber-go/zap)，按天分割，自动关联链路追踪上下文                                            |
| MySQL       | 基于 [xorm](https://xorm.io/)，连接池、读写分离、事务封装                                                               |
| Redis       | 连接池、字符串/哈希/列表/集合/有序集合操作、分布式锁、令牌桶限流                                                        |
| HTTP 客户端 | 链式 API，支持超时、重试、链路追踪自动传播                                                                              |
| 参数校验    | 基于 `go-playground/validator`，支持丰富的校验规则                                                                      |
| 定时任务    | 基于 [cron](https://github.com/robfig/cron)，秒级精度，提供便捷的时间表达                                               |
| SSE 流      | Server-Sent Events 客户端，支持流式消费                                                                                 |
| 其他        | Panic 恢复、健康检查、Pprof 性能分析、跨域中间件                                                                        |

## 安装

```bash
go get -v -t github.com/metlive/gohera
```

## 第一个项目

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/metlive/gohera"
)

func main() {
    // 初始化：解析配置 → 连接 DB/Redis → 配置日志
    engine := gohera.InitApp()

    // 注册路由
    engine.GET("/api/hello", func(c *gin.Context) {
        gohera.JsonSuccess(c, gin.H{"msg": "hello world"})
    })

    // 启动服务
    gohera.StartupService(engine)
}
```

```bash
# 本地：命令行指定部署环境
go run main.go -env=dev

# 容器 / K8s：由框架读取 APP_ENV（未传 -env 时生效）
# export APP_ENV=prod
```

## 配置

框架按以下优先级自动发现配置文件：
1. 环境变量 `APP_CONFIG_FILE` 指定的绝对路径
2. 在 `./`、`./config`、`./configs` 目录中查找 `app.toml` / `app.yaml` / `app.json` 等
3. 若上述目录恰好只有一个配置文件，自动使用它

### Nacos（可选）

`InitApp` 在加载本地 `app.*` 之后、初始化 MySQL/Redis 之前，读取 `bootstrap.yaml`：

| 行为                  | 说明                                                                |
|-----------------------|---------------------------------------------------------------------|
| `nacos.enabled=true`  | 按 `mode`（http/grpc）拉取并合并进运行时配置，并 Listen 热更新      |
| `nacos.enabled=false` | 若存在 `configs/nacos.{env}.yaml`（或 `nacos.localPath`），自动合并 |
| `failLocalFallback`   | 远程拉取失败时合并本地兜底文件                                      |
| `dataIdByEnv`         | Data ID 追加 `-{env}`（如 `my-app-dev`）                            |
| `mode`                | `http`（默认，v1 SDK / OpenAPI）或 `grpc`（v2 SDK，直连 Nacos 2.x） |
| `grpcPort`            | gRPC 端口，可选；默认 SDK 使用 `Port+1000`                          |

```yaml
# configs/bootstrap.yaml
nacos:
  enabled: false
  mode: http
  serverAddr: "http://nacos-gateway.example/tcm-api"
  prefix: "my-app"
  dataId: "my-app"
  dataIdByEnv: true
  failLocalFallback: true
  group: "DEFAULT_GROUP"
```

业务侧只需 `gohera.InitApp()`，然后使用 `gohera.Mysql` / `gohera.Redis` / `gohera.GetString`；无需自行接入 Nacos。

```toml
# config.toml

[http]
host = "0.0.0.0"
port = 8080

[zhttp]
pprof = 1              # 开启 pprof，访问 /debug/pprof

[log]
path = "/var/log/myapp"

[mysql.default]
host = "master:3306,slave1:3306,slave2:3306"
user = "root"
password = "pwd"
database = "mydb"
max_idle_conns = 10
max_open_conns = 50
max_life_time = "1h"
policy = 3             # 从库路由：0=Random 1=WeightRandom 2=RoundRobin 3=WeightRoundRobin 4=LeastConn

[redis]
address = "127.0.0.1:6379"
auth = ""
database = 0
max_idle = 10
max_active = 50
```

在代码中读取配置：

```go
host := gohera.GetString("http.host")
port := gohera.GetInt("http.port")
timeout := gohera.GetDuration("timeout")

// 更多类型
gohera.GetBool("zhttp.pprof")
gohera.GetFloat64("threshold")
gohera.GetStringSlice("allow_origins")
gohera.GetStringMap("mysql") // 获取所有 mysql 配置节点

// 检查配置是否存在
if gohera.IsSet("redis") { ... }

// 将配置反序列化到结构体
type MyConfig struct {
    Key string `mapstructure:"key"`
}
var cfg MyConfig
gohera.UnmarshalKey("myconfig", &cfg)

// 整份配置反序列化 + 热更新回调
var all AppConfig
gohera.Unmarshal(&all)
gohera.OnConfigChange(func() {
    gohera.Unmarshal(&all) // 配置变更后自行刷新本地副本
})
```

## 日志

每个请求自动关联 `trace_id` 和 `span_id`，后续日志无需手动传递。

**方式一：绑定 Context（推荐）** —— 入口处绑定一次，后续直接打印：

```go
func Handler(c *gin.Context) {
    log := gohera.Ctx(c)

    log.Info("请求开始处理")
    log.Infotf("用户 %s 登录成功", username)

    // 在 Service 层也可以直接用的包装函数
    doSomething(c)
}

func doSomething(ctx context.Context) {
    log := gohera.Ctx(ctx)          // 这里用的是 context.Context
    log.Warn("这是一个警告")
    log.Errortf("操作失败: %v", err)
}
```

**方式二：直接调用** —— 每次都传 context：

```go
gohera.Info(c, "请求开始")
gohera.Infotf(c, "用户 %s 登录", name)
gohera.Warn(c, "触发限流")
gohera.Error(c, "数据库查询失败")
gohera.Errortf(c, "查询超时: %v", err)
```

日志文件按天自动分割，路径格式：`{log.path}/{appName}_2024-01-01`，四种级别分别输出到 `server_debug.log`、`server_info.log`、`server_warn.log`、`server_error.log`。

---

如果只需要日志能力（不需要完整 InitApp），可以单独初始化：

```go
gohera.InitLogger("/var/log/myapp")                    // 仅文件输出
gohera.InitLoggerWithStdout("/var/log/myapp", true)    // 同时输出到控制台
```

## MySQL

配置文件中的第一个 host 为主库，其余为从库。读操作默认走从库，写操作走主库。

### 基础用法

```go
import (
    "github.com/metlive/gohera"
    "github.com/metlive/gohera/mysql"
)

// 通过 DB 实例直接操作
func GetUsers(c *gin.Context) ([]User, error) {
    var users []User
    // 获取名为 "default" 的数据库连接，默认走从库
    db := gohera.Mysql["default"]
    err := db.Context(c).Table("users").Where("status = ?", 1).Find(&users)
    return users, err
}

func CreateUser(c *gin.Context, user *User) error {
    db := gohera.Mysql["default"]
    _, err := db.Context(c).Table("users").Insert(user)
    return err
}
```

### 读写分离

```go
// 查询（走从库）
var items []Item
db.Context(c).Table("items").Find(&items)

// 强制走主库
engine := gohera.Mysql["default"].Engine  // 获取底层 xorm.Engine
engine.Find(&items)
```

### 事务

**自动管理（推荐）：**

```go
db := gohera.Mysql["default"]
err := db.WithTransaction(func(tx *mysql.Tx) error {
    tx.Context(c).Table("orders").Insert(order)
    tx.Context(c).Table("stock").Where("id = ?", itemID).Update(gin.H{"qty": qty - 1})
    return nil  // 返回 error 自动回滚，nil 自动提交
})
```

**手动控制：**

```go
tx, err := db.Begin()
// defer tx.Rollback()  // 安全做法
// ... 操作 ...
return tx.Commit()
```

### 多数据库

```toml
[mysql.db1]
host = "host1:3306"
database = "db1"
# ...

[mysql.db2]
host = "host2:3306"
database = "db2"
# ...
```

```go
db1 := gohera.Mysql["db1"]
db2 := gohera.Mysql["db2"]
```

## Redis

### 基础操作

```go
r := gohera.Redis

// 字符串
r.Set("key", "value")
val, _ := r.Get("key")
r.SetEx("key", "value", 60)   // 60 秒过期
r.Incr("counter")
r.Decr("counter")

// 键操作
r.Del("key")
exists, _ := r.Exists("key")
r.Expire("key", 300)          // 设置过期时间（秒）
ttl, _ := r.Ttl("key")        // 查看剩余时间
```

Hash、List、Set、Sorted Set 操作：

```go
r.HSet("user:1", "name", "张三")
r.HGet("user:1", "name")
r.HGetAll("user:1")

r.LPush("queue", "task1")
r.RPop("queue")

r.SAdd("tags", "go", "web")
r.SMembers("tags")

r.ZAdd("rank", 100, "player1")
r.ZRange("rank", 0, -1)
```

### 分布式锁

```go
lockKey := "resource:order:123"
requestID := uuid.NewString()

if ok, _ := gohera.Redis.Lock(lockKey, requestID, 10); ok {
    defer gohera.Redis.Unlock(lockKey, requestID)
    // 执行业务逻辑 ...
} else {
    // 获取锁失败，说明有其他请求正在处理
}
```

### 令牌桶限流

```go
// 每秒生成 10 个令牌，桶容量 20，本次消耗 1 个
allowed, _ := gohera.Redis.RateLimit("limit:api:login", 10, 20, 1)
if !allowed {
    gohera.JsonError(c, gohera.ErrSystem, "请求过于频繁，请稍后再试")
    return
}
```

## HTTP 客户端

链式 API，支持超时、重试、链路追踪自动传播。

```go
// GET 请求
resp := gohera.NewRequest().
    SetTimeOut(5).
    SetRetries(2).
    Get("https://api.example.com/users")

data, err := resp.String()
// 或解析 JSON
var users []User
resp.ToJSON(&users)

// POST JSON
resp := gohera.NewRequest().
    SetTimeOut(10).
    SetHeader("Authorization", "Bearer xxx").
    PostJsonCtx(c, "https://api.example.com/data", reqBody)

// POST Form
resp := gohera.NewRequest().
    PostFormCtx(c, "https://api.example.com/form", map[string]any{
        "name": "张三",
        "age":  25,
    })

// 设置 Basic Auth
resp := gohera.NewRequest().
    SetBasicAuth("user", "pass").
    Get("https://api.example.com/protected")

// 获取响应详情
statusCode := resp.GetRespStatus()
respHeader := resp.GetRespHeader()
```

链路追踪的 `x-trace-id` / `x-span-id` 会在请求中自动传播。

## 参数校验

```go
// POST JSON 校验
type CreateUserReq struct {
    Name   string `json:"name" binding:"required,min=2,max=32"`
    Email  string `json:"email" binding:"required,email"`
    Age    int    `json:"age" binding:"gte=0,lte=150"`
    Mobile string `json:"mobile" binding:"omitempty,len=11"`
}

func CreateUser(c *gin.Context) {
    var req CreateUserReq
    if err := c.ShouldBind(&req); err != nil {
        gohera.JsonError(c, gohera.ErrParam)
        return
    }
    // ...
}

// GET 参数校验
type ListReq struct {
    Page     int    `form:"page" binding:"required,gte=1"`
    PageSize int    `form:"page_size" binding:"required,gte=1,lte=100"`
    Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
}

func ListUsers(c *gin.Context) {
    var req ListReq
    if err := c.ShouldBindQuery(&req); err != nil {
        gohera.JsonError(c, gohera.ErrParam)
        return
    }
    // ...
}
```

更多校验规则参考：[go-playground/validator](https://github.com/go-playground/validator)

## 响应

统一 JSON 响应格式：`{"code": 0, "message": "", "data": ...}`

```go
// 成功
gohera.JsonSuccess(c, data)
gohera.JsonSuccess(c, gin.H{"users": users, "total": total})

// 业务错误（HTTP 200，code 为非 0）
gohera.JsonError(c, gohera.ErrParam)                          // 使用预定义错误码
gohera.JsonError(c, gohera.ErrParam, "用户名不能为空")          // 自定义错误信息

// 系统异常（HTTP 500）
gohera.JsonAbort(c, gohera.ErrSystem, "内部错误")
```

预定义错误码：

| 常量 | 值 | 说明 |
|------|------|------|
| `Success` | 0 | 操作成功 |
| `ErrParam` | 1010301 | 参数错误 |
| `ErrSystem` | 1000000 | 系统错误 |
| `ErrInternal` | 1010101 | 内部错误 |
| `ErrMysql` | 1010102 | MySQL 错误 |
| `ErrRedis` | 1010103 | Redis 错误 |
| `ErrAccessToken` | 1010201 | Token 错误 |

注册自定义错误码：

```go
gohera.SetMessage(1001, "余额不足")
gohera.JsonError(c, 1001)
```

## 中间件

框架已内置链路追踪和恢复中间件。可按需添加：

```go
engine := gohera.InitApp()

// 跨域
engine.Use(gohera.CorsContext())

// 自定义中间件
engine.Use(func(c *gin.Context) {
    // 前置处理 ...
    c.Next()
    // 后置处理 ...
})

// 路由组中间件
g := engine.Group("/api/admin")
g.Use(AdminAuth())
{
    g.GET("/dashboard", AdminDashboard)
}
```

## 定时任务

```go
job := gohera.NewJobManager()

// 每分钟
job.Command("sync-cache", func() {
    // 同步缓存 ...
}).EveryMinutes()

// 每天凌晨 3 点
job.Command("cleanup", func() {
    // 清理过期数据 ...
}).DailyAt("03:00")

// 自定义 cron 表达式
job.Command("report", func() {
    // 生成报表 ...
}).Cron("0 30 10 * * 1-5")   // 工作日 10:30

job.Start()

// 优雅关闭时停止
// job.Stop()
```

便捷时间方法：`EverySeconds()`、`EveryFiveSeconds()`、`EveryTenMinutes()`、`EveryThirtyMinutes()`、`Hourly()`、`Daily()`、`Weekly()`、`Monthly()`、`Yearly()`。

## SSE 流式消费

```go
err := gohera.StreamSSE(&gohera.StreamConfig{
    URL:     "https://api.example.com/events",
    Method:  "POST",
    Timeout: 30 * time.Second,
    Headers: map[string]string{"Authorization": "Bearer xxx"},
    Body:    jsonBody,
}, func(event *gohera.SSEEvent) error {
    fmt.Printf("[%s] %s\n", event.Event, event.Data)
    return nil  // 返回 error 可中断流读取
})
```

## 环境

部署级别由 **gohera 框架**解析，优先级：

1. 命令行显式 `-env`（本地调试）
2. 环境变量 `APP_ENV`（容器 / K8s 注入，框架直读，不进配置快照）
3. 默认 `dev`

| 环境 | 说明 | Gin Mode |
|------|------|----------|
| `dev` | 开发环境 | Debug |
| `test` | 测试环境 | Test |
| `pre` | 预发布环境 | Release |
| `prod` | 生产环境 | Release |

未传 `-env` 且未设 `APP_ENV` 时按 `dev` 启动；生产必须注入 `APP_ENV=prod` 或启动参数 `-env=prod`，否则会以 Debug 模式跑。

代码中判断环境：

```go
if gohera.IsDev() {
    // 开发环境特有逻辑
}
env := gohera.GetEnv()
```

## 健康检查

启动后自动注册，无需额外配置：

```
GET /healthz
→ {"status": 200, "env": "dev"}
```

## 异常恢复

生产环境自动捕获 Panic，记录请求详情和堆栈信息，返回统一错误响应。开发环境保留原始 Panic 便于调试。

## 安全与稳定性
>
>本章分析框架各模块的安全设计和可靠性保障，帮助你在生产环境中有信心地使用。

### 异常恢复

- **非开发环境自动捕获 Panic**，记录请求体、堆栈信息到日志，返回统一错误响应而非裸奔 500 页面
- **甄别连接断开类异常**（broken pipe / connection reset），单独标记不触发告警级日志，减少噪声
- **开发环境保留原始 Panic**，直接暴露错误便于定位问题

```go
// init.go 中的注册逻辑
if !gohera.IsDev() {
    engine.Use(gohera.HandlerRecovery(true))  // true = 记录完整堆栈
}
```

### 链路追踪与日志隔离

- 每个请求自动生成唯一 `x-trace-id` 和 `x-span-id`，请求间日志严格隔离，杜绝串行
- 日志上下文通过 `context.Context` 传递，Service 层无需依赖 Gin Context
- 跨服务 HTTP 调用自动传播追踪信息（`NewRequest` 内部处理）
- 日志不输出请求体中的敏感字段，仅记录路径、方法、状态码等元信息

### 配置安全

```
敏感信息建议通过环境变量覆盖配置文件中
的明文值，框架自动映射 APP_ 前缀的环境变量
（下划线对应配置层级，如 `APP_HTTP_PORT` → `http.port`）。
环境变量同时作用于 `Get*` 读取与 `Unmarshal*` 反序列化，
且只覆盖文件中已存在的叶子键：

  export APP_MYSQL_DEFAULT_PASSWORD=secret
  # 覆盖配置文件中 [mysql.default] 下的 password 字段
```

- 配置 Hot Reload 通过 `fsnotify` 监听文件变更，原子更新缓存
- 配置文件自动发现逻辑收敛：多文件并存的目录会报错要求明确指定，防止误加载

### 数据库

**MySQL：**
- 连接池维度防护：`MaxIdleConns` / `MaxOpenConns` / `MaxLifeTime` 均通过配置控制
- 非生产环境自动开启 SQL 日志，生产环境关闭 —— 避免泄露查询内容
- 创建连接时执行 `Ping()` 验证可用性，失败直接终止启动
- 事务 `WithTransaction` 自带 Panic 恢复：业务代码 panic 时自动回滚后重新抛出
- 读写分离：读默认走从库，写走主库，业务也可显式 `getMaster()` 强制读主

**SQL 注入防护：**
- ORM 操作（`Where`、`Find`、`Insert` 等）均通过 xorm 参数绑定，不存在拼接
- 仅当使用 `s.SQL("raw sql")` 自定义 SQL 时需自行注意参数化。推荐始终使用 `?` 占位符传参

**Redis：**
- 连接池 Dial 时完成 Auth + DB Select，借出即用
- 分布式锁：`SET NX EX` 原子加锁 + Lua 脚本 CAS 解锁，防止误删他人持有的锁
- 令牌桶限流：Lua 脚本原子执行，自动设置过期清理冷数据

### HTTP 客户端

- 默认连接池：100 最大连接，20 每 host，90s 空闲超时，30s keepalive
- 请求级超时独立控制，不受全局连接池超时影响
- 自动解压 gzip 响应
- 无论成功与否都会消费响应体并关闭（`io.Copy(io.Discard, resp.Body)` 后 `Close`），不会泄露连接
- Context 取消 / 超时时，重试循环立即退出，不会无限重试

### 并发安全

| 组件 | 保护方式 |
|------|----------|
| 配置缓存 | `atomic.Pointer` 无锁读写 |
| 数据库连接 map | `sync.RWMutex` + Double-Check Locking |
| 应用初始化 | `sync.Once` 单次执行 |
| 服务关闭状态 | `atomic.Int32` 防止重复退出 |
| 定时任务配置 | `sync.Mutex` |

### 生产环境部署建议

1. **启用参数校验**：对外接口的每个请求体都应定义 binding 规则，框架自动拒绝非法输入
2. **Redis 配置强密码**：生产环境务必设置 `auth` 字段
3. **限流**：对登录、注册等敏感接口，使用 `RateLimit` 防止暴力攻击
4. **跨域收敛**：`CorsContext` 默认允许所有来源，生产环境建议替换为自定义中间件限制具体域名
5. **优雅关闭**：框架实现了信号监听，业务侧在收到退出信号后应停止接收新请求、等待进行中请求完成

### 已知注意事项

- **配置文件明文存储**：数据库密码等敏感信息建议通过环境变量 `APP_xxx` 注入覆盖
- **自定义 SQL**：`s.SQL()` 直接执行原始 SQL，使用时务必参数化，避免拼接用户输入
- **跨域中间件**：`CorsContext` 设计为开箱即用，对安全要求较高的场景请自行替换实现
- **优雅关闭**：框架已监听退出信号，但未自动执行 `http.Server.Shutdown()`，如需要可在业务代码中自行处理

## License

MIT
