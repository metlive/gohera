# mysql / sqlite 独立包优化

**日期：** 2026-07-30  
**状态：** 已落地  
**缓存策略：** A — 同 Database / FilePath 全局复用

## 目标

- 三方可直接 `mysql.New` / `sqlite.New`，不依赖 `InitApp`。
- 与 `redis.New` API 风格对齐。
- 保持零依赖根包 `gohera`。
- `InitApp` 继续桥接配置并写入 `gohera.Mysql`。

## API

```go
db, err := mysql.New(&mysql.Config{...})
db, err := sqlite.New(&sqlite.Config{...})
db.Close()
mysql.CloseAll() / sqlite.CloseAll()
```

- mysql 保留 `InitOnce(cfg).Connect()` → 内部转 `New`（兼容）。
- 默认连接池参数；必填字段校验。
- mysql：`ShowSQL *bool` 优先；未设时 `Env=DEV|TEST` 仍开 SQL 日志。

## 非目标

不改 transaction 包装语义；不引入 gohera 配置读取。
