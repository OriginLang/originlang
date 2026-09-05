# Gateway

上层 App、CLI、桌面 UI 与外部客户端访问 OriginLang 宿主的统一入口层。

Gateway 负责把 HTTP、gRPC、WebSocket、本地 IPC 或 CLI 调用适配为统一的请求上下文，并在授权、租户解析、路由、限流、审计与可观测性之后，将调用转发给宿主的稳定服务接口。它不加载插件、不启动 Node/JRE/Python 运行时，也不实现插件业务。

## 目录规划

| 目录 | 职责 | 当前状态 |
| --- | --- | --- |
| `contracts/` | Gateway 对内使用的传输无关请求、响应、身份、租户和错误模型。 | 规划中 |
| `middleware/` | 鉴权、租户解析、限流、审计、追踪和请求策略。 | 规划中 |
| `routing/` | 系统路由与已授权插件贡献路由的解析和分发。 | 规划中 |
| `transports/` | HTTP、gRPC、WebSocket、本地 IPC、CLI 等入口协议的适配。 | 规划中 |

## 位置与边界

```text
App / CLI / Electron renderer / remote client
                    │
             gateway/transports
                    │
   gateway/middleware → gateway/routing
                    │
runtime/services + Plugin Manager + runtime/host-api
                    │
             adapters/* → plugin processes
```

- `runtime/ipc/` 定义宿主与插件之间的 JSON-RPC、stdio、TCP 等底层协议；Gateway 不复制这些协议。
- `runtime/services/` 负责权限、存储、调度等领域服务；Gateway 只编排请求策略并调用它们。
- `adapters/` 负责启动 Node、Python、JVM、Wasm 等插件载体；Gateway 不感知具体运行时。
- `apps/` 中的具体宿主决定是否部署 Gateway：服务端可公开 HTTP/gRPC，Electron 可在主进程提供仅本机可用的入口，CLI 可调用本地宿主或远端 Gateway。
- `plugins/` 通过已定义的扩展点贡献 API 路由；Gateway 只在插件已加载、已授权且路由无冲突时挂载它们。

## 统一请求模型

每种入口最终应归一为传输无关的请求：

```text
RequestContext(identity, tenant, permissions, trace, deadline)
  + Route(target, method, params)
  → Response(result | typed error)
```

统一错误、截止时间、取消信号与追踪上下文必须在 Gateway 与插件调用之间保持可传递；传输专属的状态码或连接对象不得泄漏到运行时服务接口。

## 安全原则

- Gateway 默认拒绝未认证、未授权、未绑定租户的请求。
- 插件路由的注册来自已验证 manifest 的扩展声明，不能由任意请求动态创建。
- Gateway 依据运行时的权限决策转发请求，不能因为路由已注册而绕过插件能力或租户授权。
- 本地 Electron/CLI 入口与远端 HTTP 入口应使用不同的信任边界和认证策略。

## 当前状态

本目录目前只定义结构与职责边界；尚未实现网络监听、路由器、中间件或协议适配。
