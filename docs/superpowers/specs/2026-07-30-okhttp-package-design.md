# okhttp 独立包设计

**日期：** 2026-07-30  
**状态：** 已落地  
**范围：** 使 `okhttp` 可作为三方独立 HTTP 客户端使用；同时在完整 gohera 应用中保留 service 名与 gin Context 链路能力。

---

## 1. 目标

- 场景 **C**：默认独立可用（A）；在 gohera 应用中经桥接带上 `http.service` / gin Context。
- `okhttp` **禁止** import `github.com/metlive/gohera`。
- 允许依赖：`logger`、`gin`（仅用于可选 `*gin.Context` 断言）、标准库。
- 对外用法保持 `okhttp.NewRequest()...`，三方无需 `InitApp`。

## 2. 依赖方向

```
三方 ──► okhttp ──► logger
完整应用 ──► gohera.InitApp ──► okhttp.SetDefaultService(...)
                              └► logger.Init(...)
```

## 3. API 变更

### 3.1 常量（okhttp 本地，字面量与 gohera 历史一致）

```go
const (
    FormContentType = "application/x-www-form-urlencoded"
    JsonContentType = "application/json"
    TraceIdHeader   = "x-trace-id" // 或复用同名 TraceId
    SpanIdHeader    = "x-span-id"
)
```

Trace Context key / 默认 span：使用 `logger.TraceCtx`、`logger.SpanIdDefault`。

### 3.2 Service 名

```go
func SetDefaultService(name string)     // 包级默认，线程安全
func DefaultService() string

func (h *HTTPRequest) SetService(name string) *HTTPRequest  // 单请求覆盖
```

`setReferer`：优先请求 URL；否则用 `h.service`（若空则 `DefaultService()`）；都空则不设 Referer。

### 3.3 本地 getHeader

将原 `gohera.GetHeader` 逻辑复制为 okhttp 包内私有函数，不反向依赖 gohera。

### 3.4 gin Context

保持现有 `cx.(*gin.Context)` 分支；key 改为 `logger.TraceCtx`。

## 4. gohera 桥接

`InitApp` 在配置加载后：

```go
okhttp.SetDefaultService(GetString("http.service"))
```

gohera 根包常量 `TraceCtx` / `FormContentType` 等可保留（字面量不变），供 middleware 等使用。

## 5. 非目标

- 不拆掉 gin 依赖（可选能力）。
- 不改变重试、超时、SSE 现有行为。
- 不引入新的 HTTP 抽象层。

## 6. 验证

- `go build ./...` 无环。
- `go list ./okhttp` 的 Imports 不含 `github.com/metlive/gohera`。
- 未设 service 时可完成请求；设 DefaultService 后 Referer 回退路径生效。
