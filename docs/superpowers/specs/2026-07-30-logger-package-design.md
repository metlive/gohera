# Logger 独立包设计

**日期：** 2026-07-30  
**状态：** 已落地  
**范围：** 将 `logger` 拆为可独立复用的子包，供三方只取日志能力；本仓库与完整 `gohera` 框架继续通过桥接使用。

---

## 1. 背景与问题

当前改动把日志放进 `logger` 包，意图是：

1. 三方可单独引用日志能力，不必跑 `InitApp`，不必依赖 MySQL/Redis/配置。
2. 本仓库（及使用完整 gohera 的业务）仍可统一用同一套日志。

但现状存在硬伤：

| 问题           | 说明                                                                                  |
|----------------|---------------------------------------------------------------------------------------|
| 循环依赖       | `gohera` → `logger` → `gohera`，`go build ./...` 失败                                 |
| 配置耦合       | `logger` 内调用 `gohera.GetBool` / `GetString`，独立使用时无配置或行为不确定          |
| 未导出调用     | `InitApp` 调用 `logger.initLoggerPool` / `logger.loggerConfig`，跨包不可见            |
| 空路径不安全   | `FilePath` 为空时仍拼 `./%s/server_*.log` 并创建 rotatelogs，可能写到意外目录或 panic |
| 未初始化 panic | 未 `Init` 就打日志时 `logger` 为 nil                                                  |

目标场景（已确认）： **A — 三方只拿日志能力**；同时保留日后接入完整 gohera 框架的能力。

---

## 2. 目标与非目标

### 2.1 目标

- `logger` **零依赖** `github.com/metlive/gohera`。
- 默认 / 未配置文件路径： **仅控制台**，不创建日志文件。
- 显式配置 `FilePath`： **写文件**；是否双写控制台由选项控制。
- 本仓库 `InitApp` 从配置读取后调用 `logger.Init`，行为与现网一致（可配路径、stdout）。
- 同一进程可先只用 logger，后调 `gohera.InitApp()`；`InitApp` 内再次 `Init` **覆盖**先前配置。
- 对外 API 简洁：`Init` / `Info` / `Warn` / `Error` / `Ctx` / `Trace`。

### 2.2 非目标

- 不在本方案中引入新的日志后端（如 Loki、OpenTelemetry exporter）。
- 不改变现有 JSON 文件字段名（`x_message`、`x_trace_id` 等）与按级别分文件策略。
- 不强制业务立刻去掉对 `gohera.TraceCtx` 等常量的引用（可提供兼容别名）。
- 不在本方案中做多实例 / 按模块多个 `*zap.Logger`（仍为包级单例）。

---

## 3. 依赖方向

```
                    ┌─────────────┐
   仅日志业务 ──────►│   logger    │
                    └─────────────┘
                           ▲
                           │ Init(Options)
                    ┌──────┴──────┐
   完整框架业务 ────►│   gohera    │──► mysql / redis / gin ...
                    └─────────────┘
```

- **允许：** `gohera` import `logger`；业务同时 import 两者。
- **禁止：** `logger` import `gohera`。
- 颜色、`TraceCtx`、本地工具等从 gohera 迁入 logger，或在 logger 内自实现。

---

## 4. 公开 API

### 4.1 Options

```go
package logger

// Options 控制日志输出目标与格式。零值经 Init 规范化后的语义见「默认与规范化」。
type Options struct {
    // FilePath 日志目录。空字符串：不写任何日志文件。
    // 非空时用 filepath.Join(FilePath, "server_{level}.log")，支持绝对/相对路径。
    FilePath string

    // EnableStdout 是否输出到控制台。
    // 使用指针以区分「未设置」与「显式 false」。
    // nil → 默认 true；非 nil → 按 *EnableStdout。
    EnableStdout *bool

    // StdoutFormat 控制台格式：
    //   "simple"   — 仅消息（默认）
    //   "detailed" — [时间] LEVEL [caller?] [traceId?] : 消息
    StdoutFormat string

    // Project 写入全局字段 x_project；空则不写或写空字符串（实现选定一种并固定）。
    Project string
}
```

辅助函数（可选，便于调用方）：

```go
func Bool(v bool) *bool { return &v }
```

### 4.2 初始化

```go
// Init 按 Options 配置（或重配）包级 logger。可重复调用；后一次覆盖前一次。
func Init(opts Options)

// InitLogger 兼容入口：logPath 非空则写文件；空路径则仅控制台。
// 等价于 Init(Options{FilePath: logPath, EnableStdout: Bool(true)})。
func InitLogger(logPath string)

// InitLoggerWithStdout / InitLoggerWithStdoutFormat 保留为薄封装，内部转 Init。
func InitLoggerWithStdout(logPath string, enableStdout bool)
func InitLoggerWithStdoutFormat(logPath string, enableStdout bool, stdoutFormat string)
```

### 4.3 打日志与 Context

保持现有签名与行为：

- `Info` / `Infotf` / `Warn` / `Warntf` / `Error` / `Errortf`
- `Ctx(ctx) *ContextLogger`
- `Trace` 结构体、`GetTraceContext`、`StartSpan`

### 4.4 常量（迁入或定义在 logger）

```go
const (
    TraceCtx      = "trace-ctx"
    SpanIdDefault = "1"
)
```

gohera 侧可保留同名常量并注明「与 logger 一致」，或改为：

```go
const TraceCtx = logger.TraceCtx // 若类型允许；字符串常量可直接复制定义避免循环
```

因字符串常量无类型依赖， **推荐两边各自定义相同字面量**，或 gohera middleware 直接使用 `logger.TraceCtx`。

---

## 5. 默认行为与规范化

`Init` 入口先规范化：

| 字段                  | 规范化规则                                         |
|-----------------------|----------------------------------------------------|
| `EnableStdout == nil` | 视为 `true`                                        |
| `StdoutFormat == ""`  | 视为 `"simple"`                                    |
| `StdoutFormat` 非法   | 回退 `"simple"`，并可向 stderr 打一行警告          |
| `FilePath`            | `strings.TrimSpace`；空则 **不** 添加任何文件 Core |

**懒初始化：**

- 若从未调用 `Init`，第一次 `Info`/`Warn`/`Error` 时自动：

  `Init(Options{EnableStdout: Bool(true)})`

  即仅控制台，不写文件，不 panic。

**文件输出条件：**

- 仅当规范化后 `FilePath != ""` 时，创建 debug/info/warn/error 四个 rotatelogs Core。
- 文件格式保持现有 JSON Encoder 与切割策略（按天、链接名、MaxAge 7 天等）。
- **路径拼接：** 完整文件路径为 `filepath.Join(FilePath, "server_{level}.log")`（再交给 rotatelogs 加日期后缀）。 **不再**无条件前缀 `./`，避免绝对路径被写成 `.//var/...` 而落到相对目录。相对 `FilePath` 仍相对进程 cwd。

**控制台：**

- `EnableStdout == true` 时追加 console Core（simple 或 detailed）。
- 颜色逻辑迁入 `logger`（原 `console_color.go`），detailed 模式使用。

**禁止：**

- 再通过 `gohera.GetBool("log.stdout")` 等读取配置。
- 在 `FilePath == ""` 时创建 `./server_*.log` 或 `./%/...` 类路径。

---

## 6. gohera 框架桥接

### 6.1 InitApp

```go
appPath := GetDefaultString("log.path", DefaultLogPath)
filePath := appPath + "/" + appName

var enableStdout *bool
if IsSet("log.stdout") {
    v := GetBool("log.stdout")
    enableStdout = &v
}

logger.Init(logger.Options{
    FilePath:     filePath,
    EnableStdout: enableStdout, // nil 时 logger 默认 true；若希望生产默认关控制台，见下节决策
    Project:      GetString("http.service"),
})
```

**与现状对齐说明：**

- 现状：控制台由 `EnableStdout` 或配置 `log.stdout` 开启；`InitApp` 未显式传 stdout 时，旧代码靠 `GetBool("log.stdout")`（缺省多为 false）。
- 独立 logger 默认 `EnableStdout=true` 更适合三方； **框架侧**若需保持「生产默认只写文件」，则 `InitApp` 应显式传入：
    - 未配置 `log.stdout` 时传 `Bool(false)`，或
    - 仅当 `GetBool("log.stdout")` / 开发环境为 true 时开启。

**决策（本方案采用）：**

- **logger 包默认：** `EnableStdout` 未设置 → `true`（三方友好）。
- **InitApp 桥接：** 以配置为准；若未设置 `log.stdout`，则传 `Bool(false)`，保持与当前「默认不打控制台、靠配置打开」接近的生产行为。开发环境若已有 `log.stdout=true` 配置则不受影响。

### 6.2 覆盖语义

`logger.Init` **可重复调用**，后一次完全替换包级 `*zap.Logger`（加锁）。

场景：业务先 `logger.Init(仅控制台)`，后 `gohera.InitApp()` → 按配置升级为文件（+可选控制台）。

### 6.3 调用方改造清单

| 位置                                                                 | 改动                                                                              |
|----------------------------------------------------------------------|-----------------------------------------------------------------------------------|
| `init.go`                                                            | 改为调用导出的 `logger.Init` / `Options`                                          |
| `middleware.go` / `okhttp` / `cron.go` / `recovery.go` / `gohera.go` | 继续 `logger.*`；Context key 改用 `logger.TraceCtx`（或保持字面量 `"trace-ctx"`） |
| `console_color.go`                                                   | 迁入 `logger`，从 gohera 删除或改为废弃转发（不可反向依赖）                       |
| `logger/logger.go`                                                   | 删除所有 `gohera.*` 引用；`Ternary` 改为本地 if / 小函数                          |

---

## 7. 三方接入示例

### 7.1 仅控制台（零配置）

```go
import "github.com/metlive/gohera/logger"

func main() {
    // 可不 Init；首次打日志自动仅控制台
    logger.Info(ctx, "hello")
}
```

### 7.2 显式仅控制台

```go
logger.Init(logger.Options{
    EnableStdout: logger.Bool(true),
    StdoutFormat: "detailed",
    Project:      "billing-worker",
})
```

### 7.3 开启文件

```go
logger.Init(logger.Options{
    FilePath:     "/var/log/trace/billing-worker",
    EnableStdout: logger.Bool(true),
    Project:      "billing-worker",
})
```

### 7.4 日后接入完整框架

```go
// 早期
logger.Init(logger.Options{EnableStdout: logger.Bool(true)})

// 后期引入 gohera
r := gohera.InitApp() // 内部再次 Init，按配置覆盖（可含 FilePath）
```

---

## 8. 错误与边界

| 场景                                          | 行为                                                                                                |
|-----------------------------------------------|-----------------------------------------------------------------------------------------------------|
| `FilePath` 非空但目录无权限 / rotatelogs 失败 | 保持现有：写 stderr + panic（框架启动失败可见）；若未来需降级，另开任务                             |
| 重复 `Init`                                   | 成功替换；旧 writer 由 GC / rotatelogs 关闭语义决定（实现时尽量 Close 旧 sink，若 rotatelogs 支持） |
| `Project` 为空                                | 全局字段 `x_project` 写空字符串，或省略该 Field；**实现固定为写空字符串**，与字段集合稳定           |
| 并发 `Init` 与打日志                          | `Init` 与读 logger 指针同一把 `loggerMu`；打日志时拿本地副本 `*zap.Logger`                          |

---

## 9. 测试要点

- 未 Init：调用 `Info` 不 panic，有控制台输出倾向（可用 zap 的观测或自定义 core 注入做单测）。
- `FilePath == ""`：不创建日志文件（可用临时目录断言无 `server_*.log`）。
- `FilePath != ""`：临时目录下生成对应级别文件（或至少 rotatelogs 创建成功）。
- `Init` 两次：第二次改 `FilePath` / stdout 后行为符合新配置。
- `go build ./...` 无 import cycle。
- 包 `logger` 的 `go list -f '{{.Imports}}'` 不含 `github.com/metlive/gohera`。

---

## 10. 迁移步骤（实现顺序）

1. 将颜色工具、`TraceCtx`/`SpanIdDefault`、本地 `ternary` 迁入/写入 `logger`，去掉对 gohera 的 import。
2. 实现 `Options`、`Init`、规范化、空路径跳过文件 Core、懒初始化。
3. 保留 `InitLogger*` 薄封装。
4. 修改 `InitApp` 桥接与 stdout 默认策略。
5. 全仓库 Context key / 引用对齐。
6. `go build ./...` + 上述测试。
7. 如有对外 README/示例，补充「仅 logger」与「完整 InitApp」两段。

---

## 11. 已确认决策摘要

| 项                  | 决策                                           |
|---------------------|------------------------------------------------|
| 拆包目的            | 三方可只拿日志（场景 A）                       |
| 方案                | 独立 Options + 默认仅控制台（方案 1）          |
| 日后用框架          | 可以；`gohera` → `logger`，InitApp 桥接        |
| 二次 Init           | InitApp 覆盖先前仅控制台配置                   |
| logger 默认 stdout  | 未设置 → true                                  |
| InitApp 默认 stdout | 未配置 `log.stdout` → 显式 false（贴近现生产） |

---

## 12. 开放项

无阻塞开放项。实现时可尽力而为、不改主结论：

- 重复 `Init` 时主动 Close 旧 rotatelogs（若库支持）。
