# gohera Nacos 配置源

## 目标

在 `InitApp` 内统一处理远端配置，业务项目不再各自实现 Nacos / OpenStores。

## 行为

1. 加载本地 `app.*`
2. 读取 `bootstrap.yaml` 的 `nacos.*`
3. 若启用：按 `mode` 拉取 → `MergeConfigMap` → Listen 热更新
4. 若未启用：存在 `configs/nacos.{env}.yaml` 则合并
5. 再初始化 MySQL / Redis

## mode

| mode           | SDK             | 适用场景                                                                     |
|----------------|-----------------|------------------------------------------------------------------------------|
| `http`（默认） | nacos-sdk-go v1 | Nacos 1.x/2.x OpenAPI、HTTP 网关（如 tcm-api）                               |
| `grpc`         | nacos-sdk-go v2 | 直连 Nacos 2.x，需开放 gRPC 口（默认 Port+1000，可用 `nacos.grpcPort` 覆盖） |

```yaml
nacos:
  enabled: true
  mode: grpc          # 或 http
  serverAddr: "http://127.0.0.1:8848"
  grpcPort: 9848      # 可选
  dataId: "my-app"
  dataIdByEnv: true
  failLocalFallback: true
```

## MySQL/Redis 热更新

连接仅在启动时建立；Nacos 变更只刷新配置缓存（如凭证类 `GetString`），不重连库。 Nacos 未包含的 key（如 mysql/redis 写在 app.yaml）在合并后仍保留。
