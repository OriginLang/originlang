# OriginLang 架构草案 · 落地可行性评估与实施计划

| 版本   | 日期         | 作者             | 状态 |
| :--- | :--------- | :------------- | :- |
| v0.1 | 2026-09-02 | OriginLang 架构组 | 草案 |

> 本文基于对现有代码库的全量审查，评估 [architecture-design-draft.md](file:///e:/code/originlang/apps/docs/architecture-design-draft.md) 中各模块的落地可行性，并给出可直接拆分执行的详细实施计划。

***

## 1. 现有代码库清单

### 1.1 源文件清单（11 个 Go 文件 + 4 个测试文件，约 1,600 行）

| 路径                                                                                          | 行数    | 职责                                                                                             | 成熟度    |
| :------------------------------------------------------------------------------------------ | :---- | :--------------------------------------------------------------------------------------------- | :----- |
| [rpc/rpc.go](file:///e:/code/originlang/src/go/host/rpc/rpc.go)                             | \~140 | JSON-RPC 2.0 Message 结构体 + 请求/响应/通知/错误构造器 + Unmarshal 辅助方法                                     | ✅ 完整可用 |
| [rpc/error.go](file:///e:/code/originlang/src/go/host/rpc/error.go)                         | \~40  | 标准 + 应用错误码 (-32001\~-32005) + NewError 构造器                                                     | ✅ 完整可用 |
| [rpc/id.go](file:///e:/code/originlang/src/go/host/rpc/id.go)                               | \~65  | ID 类型 (string/number/null) + 进程唯一 Allocator                                                    | ✅ 完整可用 |
| [transport/transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go)     | \~55  | Transport 接口 (Kind/Send/Read/Close/String) + Codec + Kind 枚举 (stdio/tcp/http)                  | ✅ 完整可用 |
| [transport/endpoint.go](file:///e:/code/originlang/src/go/host/transport/endpoint.go)       | \~200 | Endpoint: Handler 注册 + dispatch 循环 + Call/Notify + pending 响应关联 + Close                        | ✅ 完整可用 |
| [transport/stdio/stdio.go](file:///e:/code/originlang/src/go/host/transport/stdio/stdio.go) | \~90  | stdio Transport (newline-delimited JSON over io.Reader/Writer)                                 | ✅ 完整可用 |
| [transport/tcp/tcp.go](file:///e:/code/originlang/src/go/host/transport/tcp/tcp.go)         | \~150 | TCP Transport + Server (Listen/Accept) + Dial                                                  | ✅ 完整可用 |
| [plugin/plugin.go](file:///e:/code/originlang/src/go/host/plugin/plugin.go)                 | \~210 | Plugin 句柄: Spec/State 状态机 (7 态) + Capability + RegisterReply + Call/Shutdown + newID           | ✅ 完整可用 |
| [manager/manager.go](file:///e:/code/originlang/src/go/host/manager/manager.go)             | \~280 | Manager: Load (子进程 stdio) + Attach (远程 TCP) + handshake + Get/List/Call/Shutdown/StopAll/Close | ✅ 完整可用 |
| [manager/errors.go](file:///e:/code/originlang/src/go/host/manager/errors.go)               | \~65  | NotFoundError + toRPCError 转换                                                                  | ✅ 完整可用 |
| [sdk/sdk.go](file:///e:/code/originlang/src/go/host/sdk/sdk.go)                             | \~95  | 插件侧 Go SDK: Serve() + register/ping/shutdown 处理器 + 信号处理                                        | ✅ 完整可用 |
| tests/go/host/rpc/rpc\_test.go                                                              | \~170 | 单元测试: 请求/响应/通知/错误构造 + Unmarshal + Allocator 唯一性 + 往返序列化                                        | ✅ 全部通过 |
| tests/go/host/transport/endpoint\_test.go                                                   | \~190 | 单元测试: 请求-响应 + 通知 + MethodNotFound + 10 并发调用                                                    | ✅ 全部通过 |
| tests/go/host/transport/stdio/stdio\_test.go                                                | \~125 | 单元测试: 单向发送 + 双向往返 + Kind                                                                       | ✅ 全部通过 |
| tests/go/host/transport/tcp/tcp\_test.go                                                    | \~95  | 单元测试: TCP 往返 + Kind                                                                            | ✅ 全部通过 |

### 1.2 Bazel 构建配置

| 文件                                                                  | 内容                                                                                                    |
| :------------------------------------------------------------------ | :---------------------------------------------------------------------------------------------------- |
| [MODULE.bazel](file:///e:/code/originlang/MODULE.bazel)             | Bazel 9 模块: rules\_cc 0.2.22, rules\_java 9.9.0, rules\_rust 0.73.0, rules\_python 1.9.2; TS 规则已注释待启用 |
| [src/go/BUILD.bazel](file:///e:/code/originlang/src/go/BUILD.bazel) | gazelle + go\_library "host" 聚合 6 个子包                                                                 |
| 各子包 BUILD.bazel                                                     | 每个 Go 包独立的 go\_library 目标                                                                             |
| [go\_deps.bzl](file:///e:/code/originlang/src/go/go_deps.bzl)       | Gazelle 依赖占位符（待 `bazel mod tidy` 填充）                                                                  |

### 1.3 空占位目录（仅有 .gitkeep）

```
src/cpp/core/    src/java/core/    src/python/core/
src/rust/core/   src/ts/core/     apps/BUILD  services/BUILD
tools/BUILD      tests/BUILD      third_party/BUILD
```

### 1.4 关键设计特征（从代码中提取的已有约定）

1. **零第三方依赖**：全部 Go 代码只用标准库（`encoding/json`, `net`, `os/exec`, `sync`, `context`），无需 `go mod tidy` 拉依赖
2. **json.RawMessage 透传**：Message 的 Params/Result 使用 `json.RawMessage` 避免 lossy re-marshalling，这对跨语言互操作至关重要
3. **Transport 对称设计**：同一接口同时被宿主和插件实现，使远程/本地插件统一处理
4. **Manager 的 Load/Attach 双入口**：已支持子进程 (stdio) 和远程 TCP 两种部署模式
5. **SDK 的 Serve 模式**：Go 插件只需 `sdk.Serve(ctx, Options{...})` 一行即可成为合规插件

***

## 2. 架构草案逐模块可行性评估

### 2.1 评估矩阵

| 架构层          | 模块                          | 可行性     | 现有基础                          | 工作量  | 风险                                  |
| :----------- | :-------------------------- | :------ | :---------------------------- | :--- | :---------------------------------- |
| **L3 协议**    | JSON-RPC 2.0 核心             | ✅ 已完成   | rpc/ 3 文件 100% 覆盖             | 0    | 无                                   |
| **L3 协议**    | 错误码扩展 (-32006\~-32007)      | ✅ 直接改   | error.go 加 2 个常量              | 0.5h | 无                                   |
| **L3 协议**    | Manifest Schema 规范          | ✅ 全新    | 无现有代码                         | 8h   | 需 JSON Schema 设计评审                  |
| **L3 传输**    | stdio Transport             | ✅ 已完成   | stdio/ 100%                   | 0    | 无                                   |
| **L3 传输**    | TCP Transport               | ✅ 已完成   | tcp/ 100%                     | 0    | 无                                   |
| **L3 传输**    | HTTP Transport              | ⚠️ 中等   | 无，但 transport.go 已预留 KindHTTP | 16h  | 需 HTTP/JSON-RPC 双向（server+client）   |
| **L3 传输**    | Wasm Transport (in-process) | ⚠️ 中等   | 无                             | 24h  | 需嵌入 Wasmtime，Rust FFI 桥接            |
| **L3 传输**    | Native Transport (dlopen)   | ⚠️ 中等   | 无                             | 16h  | 需 C ABI 约定 + 信号安全处理                 |
| **L3 传输**    | Inproc Transport            | ✅ 低难度   | 无                             | 4h   | Go channel + 接口适配                   |
| **L2 内核**    | Plugin Manager              | ✅ 已完成   | manager/ 280 行                | 0    | 无                                   |
| **L2 内核**    | Plugin State Machine 扩展     | ✅ 低难度   | plugin.go 已有 7 态              | 2h   | 加 StateUnloading/StateDraining      |
| **L2 内核**    | 实例池 (Instance Pool)         | ⚠️ 中等   | 无，但 Manager 的 plugins map 可扩展 | 16h  | 需池化语义 + 取/还/驱逐策略                    |
| **L2 内核**    | 健康检查 (ping heartbeat)       | ✅ 低难度   | plugin.go 已有 MethodPing 常量    | 4h   | 仅需定时器 + 失败计数                        |
| **L2 内核**    | Extension Point Registry    | ✅ 全新可做  | 无                             | 16h  | 核心差异化模块，需接口设计                       |
| **L2 内核**    | Event Bus (同步)              | ✅ 低难度   | 无                             | 8h   | Go channel + 通配符匹配                  |
| **L2 内核**    | Event Bus (异步/NATS)         | ⚠️ 中等   | 无                             | 16h  | 需引入 NATS 客户端依赖                      |
| **L2 内核**    | Security Engine             | ⚠️ 中等偏难 | 无                             | 24h  | 三层权限模型设计 + 运行时拦截器                   |
| **L2 内核**    | Resource Quota (Go 层)       | ⚠️ 中等   | 无                             | 12h  | 令牌桶 + context 超时                    |
| **L2 内核**    | Resource Quota (Rust FFI)   | ⚠️ 中等偏难 | 无                             | 24h  | Rust 令牌桶 + C ABI + Go FFI 绑定        |
| **L2 内核**    | Observability Collector     | ✅ 低难度   | 无                             | 12h  | Go 原生 prometheus + otel SDK         |
| **L2 内核**    | Kernel Facade (集成层)         | ✅ 低难度   | manager.go 是事实 Facade 雏形      | 8h   | 聚合上述子系统                             |
| **L4 SDK**   | Go Host SDK                 | ✅ 已完成   | sdk/ + manager/ = 宿主+插件双侧     | 2h   | 封装 Facade 即可                        |
| **L4 SDK**   | Go Plugin PDK               | ✅ 已完成   | sdk/ Serve()                  | 0    | 无                                   |
| **L4 SDK**   | Rust Host SDK               | ⚠️ 中等偏难 | 无 (src/rust/core/.gitkeep)    | 40h  | 需 Rust→C ABI→Go 内核或 Rust 独立实现       |
| **L4 SDK**   | Rust Plugin PDK             | ⚠️ 中等   | 无                             | 24h  | 仿 Go sdk.Serve() 结构                 |
| **L4 SDK**   | Java Host SDK               | ⚠️ 中等偏难 | 无 (src/java/core/.gitkeep)    | 32h  | Java 22 FFM 调用 liboriginlang        |
| **L4 SDK**   | Java Plugin PDK             | ⚠️ 中等   | 无                             | 16h  | stdio JSON-RPC + 子进程                |
| **L4 SDK**   | Python Host SDK             | ⚠️ 中等偏难 | 无 (src/python/core/.gitkeep)  | 24h  | cffi 绑定                             |
| **L4 SDK**   | Python Plugin PDK           | ⚠️ 中等   | 无                             | 12h  | stdio JSON-RPC + 子进程                |
| **L4 SDK**   | C++ Host SDK                | ⚠️ 中等   | 无 (src/cpp/core/.gitkeep)     | 24h  | 头文件 + 链接 liboriginlang              |
| **L4 SDK**   | C++ Plugin PDK              | ⚠️ 中等   | 无                             | 16h  | stdio JSON-RPC                      |
| **L4 SDK**   | TS Host SDK                 | ⚠️ 中等   | 无 (src/ts/core/.gitkeep)      | 20h  | 纯 JS 重实现或 node:ffi                  |
| **L4 SDK**   | TS Plugin PDK               | ⚠️ 中等   | 无                             | 12h  | stdio JSON-RPC + Node 子进程           |
| **L5 运行时**   | Wasm 沙箱加载器                  | ⚠️ 中等偏难 | 无                             | 32h  | Wasmtime 嵌入 + WIT 解析 + Component 加载 |
| **L5 运行时**   | 子进程加载器                      | ✅ 已完成   | manager.Load()                | 0    | 无                                   |
| **L5 运行时**   | 原生动态库加载器                    | ⚠️ 中等   | 无                             | 16h  | dlopen + C ABI 符号解析                 |
| **L5 运行时**   | 远程 HTTP 加载器                 | ⚠️ 低难度  | manager.Attach() + tcp        | 8h   | HTTP Transport 的上层封装                |
| **L6 UI 扩展** | UI Extension Registry       | ✅ 可做    | 无                             | 12h  | 扩展点注册表的子集                           |
| **L6 UI 扩展** | Web Components 规范           | ⚠️ 中等   | 无                             | 24h  | 需 TS 前端工程化                          |
| **L6 UI 扩展** | Module Federation 集成        | ⚠️ 中等   | 无                             | 16h  | Webpack 5 MF 配置模板                   |
| **L6 UI 扩展** | React/Vue 适配器               | ⚠️ 中等   | 无                             | 16h  | 各一个 wrapper 包                       |
| **L7 治理**    | OCI Registry 适配             | ⚠️ 中等偏难 | 无                             | 24h  | ORAS 库 + cosign 签名                  |
| **L7 治理**    | 插件市场 Web                    | ⚠️ 中等偏难 | 无                             | 40h  | 全栈 Web 应用                           |
| **L7 治理**    | 依赖解析器 (SAT)                 | ⚠️ 难    | 无                             | 32h  | SAT 求解器或简化拓扑排序                      |
| **L7 治理**    | 多租户控制台                      | ⚠️ 中等偏难 | 无                             | 40h  | 全栈 Web 应用                           |
| **L7 治理**    | K8s Operator                | ⚠️ 中等偏难 | 无                             | 32h  | CRD + controller-runtime            |
| **L1 构建**    | Bazel 自定义规则                 | ⚠️ 中等   | MODULE.bazel 已配 4 语言规则        | 16h  | originlang\_plugin.bzl Starlark     |
| **L1 构建**    | `ol` CLI 工具                 | ⚠️ 中等   | 无                             | 24h  | cobra/urfave-cli + 子命令              |
| **L1 构建**    | Contract Test 框架            | ✅ 可做    | 无                             | 16h  | golden JSON 用例 + 多语言 runner         |

### 2.2 可行性等级汇总

| 等级      | 模块数 | 占比  | 说明                    |
| :------ | :-- | :-- | :-------------------- |
| ✅ 已完成   | 7   | 13% | 直接可用的现有代码             |
| ✅ 低难度   | 10  | 18% | 基于 Go 标准库或扩展现有代码，<8h  |
| ⚠️ 中等   | 22  | 40% | 需新代码但路径清晰，8-24h       |
| ⚠️ 中等偏难 | 13  | 24% | 需跨语言/新依赖/较大设计，24-40h  |
| ⚠️ 难    | 3   | 5%  | 需深入领域知识（SAT 求解器等）>32h |

### 2.3 关键风险

| #  | 风险                           | 影响                       | 缓解                                          |
| :- | :--------------------------- | :----------------------- | :------------------------------------------ |
| K1 | **Wasmtime 嵌入 Go 的 FFI 复杂度** | L5 Wasm 加载器是核心差异化能力      | 先用 `wasmtime-go` 绑定包做 MVP，不自己写 FFI          |
| K2 | **六语言 SDK 同步发布的维护成本**        | 12 套 SDK 需要 API 一致性      | 定义 C ABI 为唯一真源；SDK 都是薄绑定；Contract Test 自动验证 |
| K3 | **UI 扩展层跨框架兼容性**             | React/Vue/Angular 生态差异   | Web Components 标准为载体，MF 为懒加载补充              |
| K4 | **原生动态库 (dlopen) 的安全风险**     | 段错误拉垮宿主进程                | 默认禁用，需管理员白名单；文档明确推荐顺序                       |
| K5 | **SAT 依赖求解器复杂度**             | M3 才需要，但设计影响 Manifest 格式 | M1 先用拓扑排序 + semver 简单匹配，M3 再升级 SAT          |

***

## 3. 详细实施计划

### 3.1 M1 · 内核闭环 (0\~3 个月)

**目标**：基于现有代码，扩展到 4 种传输 + Manifest 解析 + 扩展点注册表 + Security Engine v1 + `ol` CLI + Go/TS SDK/PDK，跑通一个示例应用。

#### Sprint 1.1 — 协议扩展 + Manifest（第 1-2 周）

| 任务 ID  | 任务                                                                                                                                                                                                                                                                          | 依赖     | 产出                 | 预估   | 优先级 |
| :----- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :----- | :----------------- | :--- | :-- |
| T1.1.1 | 扩展错误码：在 [error.go](file:///e:/code/originlang/src/go/host/rpc/error.go) 中新增 `ErrCodeUnauthorized(-32006)` `ErrCodeMethodForbidden(-32007)` `ErrCodeTenantIsolated(-32008)` `ErrCodeQuotaExceeded(-32009)` `ErrCodeCircuitBroken(-32010)` `ErrCodeDependencyMissing(-32011)` | 无      | error.go 修改        | 0.5h | P0  |
| T1.1.2 | 定义 Manifest Go 结构体：`src/go/host/manifest/manifest.go`，含 `Manifest` `Artifact` `Dependency` `Capability` `Extension` 结构体 + `ParseFile(path) (*Manifest, error)` + `Validate()` 方法                                                                                            | T1.1.1 | manifest/ 包        | 6h   | P0  |
| T1.1.3 | 编写 Manifest JSON Schema 文件：`apps/docs/manifest-v1.schema.json`（对应草案 §3.3 的 Manifest 示例）                                                                                                                                                                                     | T1.1.2 | schema 文件          | 2h   | P1  |
| T1.1.4 | Manifest 单元测试：解析真实 manifest.json、校验必填字段、依赖版本解析（semver 范围匹配，简易实现）                                                                                                                                                                                                            | T1.1.2 | manifest\_test.go  | 4h   | P0  |
| T1.1.5 | 依赖解析器 v1（拓扑排序 + semver 范围匹配 `>=`, `~`, `^`，不含 SAT）：`src/go/host/manifest/resolver.go`                                                                                                                                                                                       | T1.1.2 | resolver.go + test | 6h   | P1  |

#### Sprint 1.2 — 新增 2 种传输 + 传输选择器（第 3-4 周）

| 任务 ID  | 任务                                                                                                                                       | 依赖            | 产出                 | 预估   | 优先级 |
| :----- | :--------------------------------------------------------------------------------------------------------------------------------------- | :------------ | :----------------- | :--- | :-- |
| T1.2.1 | Inproc Transport：`src/go/host/transport/inproc/inproc.go`，基于 Go channel 的同进程直接调用，实现 Transport 接口                                         | 无             | inproc.go + test   | 4h   | P0  |
| T1.2.2 | HTTP Transport：`src/go/host/transport/http/http.go`，客户端侧用 `net/http` 发 POST JSON-RPC；服务端侧用 `http.Handler` 接收，newline 或 Content-Length 分帧 | 无             | http.go + test     | 16h  | P1  |
| T1.2.3 | 扩展 [transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go) 的 Kind 枚举：新增 `KindWasm` `KindNative` `KindInproc`       | 无             | transport.go 修改    | 0.5h | P0  |
| T1.2.4 | 传输选择器：`src/go/host/transport/selector.go`，根据 Manifest.artifacts + 运行时能力自动选择最优 Transport                                                  | T1.2.1 T1.2.3 | selector.go + test | 4h   | P1  |

#### Sprint 1.3 — 扩展点注册表 + 实例池（第 5-6 周）

| 任务 ID  | 任务                                                                                                                         | 依赖            | 产出              | 预估  | 优先级 |
| :----- | :------------------------------------------------------------------------------------------------------------------------- | :------------ | :-------------- | :-- | :-- |
| T1.3.1 | ExtensionPoint + Extension 结构体：`src/go/host/extpoint/extpoint.go`，含 `Registry` 类型（Register/Unregister/Query 方法）            | 无             | extpoint.go     | 4h  | P0  |
| T1.3.2 | 内置 6 个后端扩展点定义：`api.handler` `pipeline.hook` `data.source` `auth.provider` `job.scheduler` + `originlang.host_call`（宿主能力声明） | T1.3.1        | builtin.go      | 2h  | P0  |
| T1.3.3 | Manifest → 扩展点自动注册：Manager 在 handshake 成功后，解析 Manifest.extensions，调用 Registry.Register                                     | T1.1.2 T1.3.1 | manager.go 扩展   | 4h  | P0  |
| T1.3.4 | 扩展点查询 API：`Registry.Query(pointID) []Extension` + 按 Scope (Global/Tenant/User) 过滤                                          | T1.3.1        | query.go + test | 4h  | P0  |
| T1.3.5 | 插件实例池：`src/go/host/pool/pool.go`，基于 Manifest 配置的 `pool_size`，对 wasm/native/inproc 载体管理多实例，支持 Get/Return/Evict              | 无             | pool.go + test  | 12h | P1  |

#### Sprint 1.4 — Security Engine v1 + 健康检查（第 7-8 周）

| 任务 ID  | 任务                                                                                                            | 依赖     | 产出                | 预估 | 优先级 |
| :----- | :------------------------------------------------------------------------------------------------------------ | :----- | :---------------- | :- | :-- |
| T1.4.1 | Security Engine 核心：`src/go/host/security/engine.go`，实现三层拦截：① 方法在 Capability 声明中？ ② 租户 Grant 允许？ ③ 用户 RBAC 允许？ | T1.1.1 | engine.go         | 8h | P0  |
| T1.4.2 | 租户 Grant 存储接口：`src/go/host/security/store.go`（接口定义 + 内存实现 `MemoryStore`）                                      | T1.4.1 | store.go          | 4h | P0  |
| T1.4.3 | Security Engine 集成到 Manager.Call：在 `manager.Call()` 中插入 Security Engine.Check 三层拦截，失败返回对应错误码                  | T1.4.1 | manager.go 扩展     | 4h | P0  |
| T1.4.4 | 健康检查：`src/go/host/health/checker.go`，周期性调用 `plugin.ping`，连续 3 次失败标记 StateFailed + 触发回调                        | 无      | checker.go + test | 4h | P1  |
| T1.4.5 | Resource Quota v1 (Go 原生)：`src/go/host/quota/quota.go`，令牌桶限速 + context 超时 + 并发计数（不含 Rust FFI）                 | T1.1.1 | quota.go + test   | 8h | P1  |
| T1.4.6 | Quota 集成到 Manager.Call：每次 Call 前检查 Quota，超限返回 `ErrCodeQuotaExceeded`                                          | T1.4.5 | manager.go 扩展     | 2h | P1  |

#### Sprint 1.5 — CLI 工具 + 示例应用（第 9-10 周）

| 任务 ID  | 任务                                                                                              | 依赖                | 产出              | 预估  | 优先级 |
| :----- | :---------------------------------------------------------------------------------------------- | :---------------- | :-------------- | :-- | :-- |
| T1.5.1 | `ol` CLI 骨架：`tools/cli/main.go`，子命令框架（用 `flag` 标准库，不引入 cobra 以保持零依赖）                            | 无                 | main.go         | 2h  | P0  |
| T1.5.2 | `ol plugin init`：交互式生成插件脚手架（Manifest + 插件源码模板 + BUILD.bazel）                                    | T1.5.1 T1.1.2     | init.go         | 6h  | P0  |
| T1.5.3 | `ol plugin build`：调用 Bazel 构建插件包                                                                | T1.5.1            | build.go        | 4h  | P1  |
| T1.5.4 | `ol plugin test`：启动 Manager + 加载插件 + 执行 ping + 验证 Manifest 声明的方法                                | T1.5.1            | test\_cmd.go    | 4h  | P1  |
| T1.5.5 | `ol run`：启动一个最小宿主进程，加载指定插件包，暴露 JSON-RPC 端口供调试                                                   | T1.5.1            | run.go          | 6h  | P1  |
| T1.5.6 | 示例应用 TODO-SaaS：`apps/todo-saas/`，一个 Go 宿主 + 3 个插件（Go stdio + TS stdio + Go inproc）+ 1 个菜单 UI 扩展 | T1.5.2\~T1.5.5 全部 | apps/todo-saas/ | 16h | P0  |

#### Sprint 1.6 — TS SDK/PDK + 端到端测试（第 11-12 周）

| 任务 ID  | 任务                                                                                                                                  | 依赖            | 产出           | 预估  | 优先级 |
| :----- | :---------------------------------------------------------------------------------------------------------------------------------- | :------------ | :----------- | :-- | :-- |
| T1.6.1 | TS Host SDK：`src/ts/core/host/`，JSON-RPC 2.0 编解码 + Transport 抽象 (stdio via child\_process + TCP via net + HTTP via fetch) + Manager | 无             | ts/host/     | 12h | P0  |
| T1.6.2 | TS Plugin PDK：`src/ts/core/pdk/`，Serve() 函数 (stdio JSON-RPC + register/ping/shutdown 处理)                                            | 无             | ts/pdk/      | 6h  | P0  |
| T1.6.3 | TS 插件示例：`apps/todo-saas/plugins/ts-plugin/`，一个 TypeScript stdio 插件                                                                  | T1.6.2        | ts plugin    | 4h  | P0  |
| T1.6.4 | 端到端测试：Go 宿主加载 Go 插件 (stdio) + TS 插件 (stdio) + Go 插件 (inproc)，验证调用 + Manifest 解析 + 扩展点注册 + 健康检查                                      | T1.5.6 T1.6.3 | e2e\_test.go | 8h  | P0  |
| T1.6.5 | Bazel TS 规则启用：取消注释 MODULE.bazel 中的 aspect\_rules\_js/rules\_ts                                                                      | 无             | MODULE.bazel | 1h  | P1  |
| T1.6.6 | Contract Test 框架 v1：`tests/contract/`，定义一组 golden JSON RPC 请求-响应对，Go + TS SDK 都必须通过                                                 | T1.6.1 T1.6.2 | contract/    | 8h  | P1  |

**M1 交付物汇总：**

- 扩展后的 Go 内核：6 种传输 + Manifest + 扩展点 + Security v1 + Quota v1 + 健康检查

- `ol` CLI v1 (init/build/test/run)

- Go Host SDK + Go Plugin PDK（现有已基本完成）

- TS Host SDK + TS Plugin PDK

- 示例应用 TODO-SaaS（3 插件 + 1 UI 扩展）

- 端到端测试 + Contract Test 框架

**M1 完成标志**：`ol plugin init` 生成一个 Go 插件 → `ol build` 构建 → `ol run` 启动宿主 → Go 宿主通过 JSON-RPC 调用插件方法 → 返回正确结果；同时 TS 插件也能被同一 Go 宿主加载和调用。

***

### 3.2 M2 · 六语言齐平 + 可观测性（3\~6 个月）

**目标**：Rust/Java/Python/C++ 四种 Host SDK + Plugin PDK 全部 v1.0；可观测性（Prometheus + OpenTelemetry）；UI 扩展 SDK v1；热升级/回滚。

#### Sprint 2.1 — Rust SDK/PDK + Rust FFI 层（第 1-4 周）

| 任务 ID  | 任务                                                                                                                                 | 依赖     | 产出                   | 预估  | 优先级 |
| :----- | :--------------------------------------------------------------------------------------------------------------------------------- | :----- | :------------------- | :-- | :-- |
| T2.1.1 | Rust C ABI 层：`src/rust/core/c_abi/`，定义 `originlang_*` C 函数（init/call/shutdown/register\_handler），导出 `liboriginlang.so/.dll/.dylib` | 无      | c\_abi/              | 16h | P0  |
| T2.1.2 | Rust Host SDK：`src/rust/core/sdk/`，封装 C ABI 调用 + Rust idiomatic API                                                                | T2.1.1 | sdk/                 | 12h | P0  |
| T2.1.3 | Rust Plugin PDK：`src/rust/core/pdk/`，stdio JSON-RPC + register/ping/shutdown（仿 Go sdk.Serve 结构）                                    | 无      | pdk/                 | 8h  | P0  |
| T2.1.4 | Rust 插件示例：`apps/todo-saas/plugins/rust-plugin/`，一个 Rust stdio 插件                                                                   | T2.1.3 | rust plugin          | 4h  | P0  |
| T2.1.5 | Rust Contract Test 适配：Rust SDK 跑通 `tests/contract/` golden 用例                                                                      | T2.1.2 | rust\_contract\_test | 4h  | P0  |
| T2.1.6 | Rust Bazel 构建：`src/rust/core/BUILD.bazel`，rules\_rust 配置 cdylib + test                                                             | 无      | BUILD.bazel          | 4h  | P0  |

#### Sprint 2.2 — Java/Python/C++ SDK/PDK（第 5-10 周，三语言并行）

| 任务 ID  | 任务                                                                                                | 依赖                   | 产出             | 预估  | 优先级 |
| :----- | :------------------------------------------------------------------------------------------------ | :------------------- | :------------- | :-- | :-- |
| T2.2.1 | Java Host SDK：`src/java/core/sdk/`，Java 22 FFM (Foreign Function & Memory API) 调用 `liboriginlang` | T2.1.1               | sdk/           | 12h | P0  |
| T2.2.2 | Java Plugin PDK：`src/java/core/pdk/`，stdio JSON-RPC + 子进程 JVM                                     | 无                    | pdk/           | 8h  | P0  |
| T2.2.3 | Python Host SDK：`src/python/core/sdk/`，cffi 绑定 `liboriginlang`                                    | T2.1.1               | sdk/           | 10h | P0  |
| T2.2.4 | Python Plugin PDK：`src/python/core/pdk/`，stdio JSON-RPC + 子进程 CPython                             | 无                    | pdk/           | 6h  | P0  |
| T2.2.5 | C++ Host SDK：`src/cpp/core/sdk/`，头文件 `originlang.h` + 链接 `liboriginlang`                          | T2.1.1               | sdk/ + .h      | 10h | P0  |
| T2.2.6 | C++ Plugin PDK：`src/cpp/core/pdk/`，stdio JSON-RPC + header-only                                   | 无                    | pdk/           | 8h  | P0  |
| T2.2.7 | 三语言插件示例：Java/Python/C++ 各一个 stdio 插件接入 TODO-SaaS                                                  | T2.2.2 T2.2.4 T2.2.6 | 3 plugins      | 6h  | P0  |
| T2.2.8 | 三语言 Contract Test 适配                                                                              | T2.2.1 T2.2.3 T2.2.5 | contract tests | 6h  | P0  |
| T2.2.9 | 三语言 Bazel 构建：rules\_java/rules\_python/rules\_cc 配置                                               | 无                    | 3×BUILD.bazel  | 6h  | P1  |

#### Sprint 2.3 — 可观测性 + 热升级（第 11-12 周）

| 任务 ID  | 任务                                                                                                                        | 依赖     | 产出          | 预估  | 优先级 |
| :----- | :------------------------------------------------------------------------------------------------------------------------ | :----- | :---------- | :-- | :-- |
| T2.3.1 | Prometheus 指标：`src/go/host/observ/metrics.go`，自动采集 `plugin_call_duration_ms` / `plugin_call_total` / `plugin_cpu_seconds` | 无      | metrics.go  | 4h  | P0  |
| T2.3.2 | OpenTelemetry 追踪：`src/go/host/observ/trace.go`，每次 Call 自动注入 `traceparent` 到 RPC params，跨插件全链路                             | 无      | trace.go    | 6h  | P0  |
| T2.3.3 | 结构化日志：`src/go/host/observ/log.go`，统一字段 plugin\_id/tenant\_id/method/req\_id/duration\_ms/error\_code                      | 无      | log.go      | 4h  | P0  |
| T2.3.4 | 热升级管理器：`src/go/host/hotreload/manager.go`，新旧版本并存 → 流量切分 → 观测 → 自动回滚                                                       | T1.3.5 | hotreload/  | 12h | P1  |
| T2.3.5 | 慢插件检测：`src/go/host/observ/profiler.go`，P95 > threshold×3 时输出 CPU Profile                                                  | T2.3.1 | profiler.go | 6h  | P2  |

#### Sprint 2.4 — UI 扩展 SDK v1（第 11-12 周并行）

| 任务 ID  | 任务                                                                                                 | 依赖            | 产出                  | 预估 | 优先级 |
| :----- | :------------------------------------------------------------------------------------------------- | :------------ | :------------------ | :- | :-- |
| T2.4.1 | UI Extension Registry (TS)：`src/ts/core/ui/registry.ts`，从宿主内核拉取 `extensions["originlang.ui.*"]` 列表 | T1.6.1        | registry.ts         | 4h | P0  |
| T2.4.2 | UI SDK 核心：`src/ts/core/ui/host.ts`，`window.OriginLangHost.call(method, params)` → JSON-RPC → 后端    | T1.6.1        | host.ts             | 6h | P0  |
| T2.4.3 | Web Component 适配：`src/ts/core/ui/component.ts`，`defineCustomElement(name, factory)` 帮助函数           | 无             | component.ts        | 4h | P0  |
| T2.4.4 | 4 个 UI 扩展点实现：`ui.menu` + `ui.route` + `ui.component` + `ui.setting_panel`                          | T2.4.1        | extensions/         | 8h | P0  |
| T2.4.5 | React 适配器：`src/ts/core/ui/react-adapter.ts`，把 Web Component 包成 React 组件                            | T2.4.3        | react-adapter.ts    | 4h | P1  |
| T2.4.6 | TODO-SaaS 前端：用 React + UI SDK 渲染插件贡献的菜单/路由/组件                                                      | T2.4.4 T2.4.5 | apps/todo-saas/web/ | 8h | P0  |

**M2 完成标志**：CI 中 6 语言 × 2 种传输 (stdio + inproc/http) 的 Contract Test 全绿；P95 延迟、错误率在 Prometheus + Grafana 大盘可见；10 个插件同时热升级无中断请求。

***

### 3.3 M3 · 多租户 + 插件市场（6\~9 个月）

#### Sprint 3.1 — OCI 分发 + 签名

| 任务 ID  | 任务                                                                              | 依赖            | 产出      | 预估 | 优先级 |
| :----- | :------------------------------------------------------------------------------ | :------------ | :------ | :- | :-- |
| T3.1.1 | OCI Artifact 定义：插件包 = OCI layer (wasm + manifest + ui)，`src/go/host/pkg/oci.go` | T1.1.2        | oci.go  | 8h | P0  |
| T3.1.2 | `ol plugin push/pull`：调用 ORAS 库推送/拉取插件包                                         | T3.1.1 T1.5.1 | cli 扩展  | 8h | P0  |
| T3.1.3 | cosign 签名验证：`src/go/host/pkg/sign.go`                                           | T3.1.1        | sign.go | 6h | P1  |

#### Sprint 3.2 — 多租户

| 任务 ID  | 任务                                                                                                   | 依赖            | 产出                 | 预估 | 优先级 |
| :----- | :--------------------------------------------------------------------------------------------------- | :------------ | :----------------- | :- | :-- |
| T3.2.1 | 多租户 Context：`src/go/host/tenant/context.go`，`TenantContext{ID, UserID, Permissions}` 贯穿 Manager.Call | T1.4.1        | context.go         | 4h | P0  |
| T3.2.2 | 租户-插件授权 Store：扩展 security/store.go，支持按租户查询已授权插件列表 + 配额配置                                             | T1.4.2        | store.go 扩展        | 6h | P0  |
| T3.2.3 | 多租户隔离测试：3 租户同时使用，验证可见性 + 调用隔离 + 配额独立                                                                 | T3.2.1 T3.2.2 | isolation\_test.go | 8h | P0  |

#### Sprint 3.3 — 插件市场 + 控制台

| 任务 ID  | 任务                                                      | 依赖     | 产出          | 预估  | 优先级 |
| :----- | :------------------------------------------------------ | :----- | :---------- | :-- | :-- |
| T3.3.1 | 插件市场 API 后端：`services/market/`，插件 CRUD + 搜索 + 评级 + 安装统计 | T3.1.1 | market API  | 16h | P0  |
| T3.3.2 | 插件市场 Web 前端：`apps/market-web/`，React + 搜索 + 详情页 + 安装按钮  | T3.3.1 | market-web/ | 16h | P0  |
| T3.3.3 | 多租户控制台：`apps/console/`，租户管理员：授权/禁用插件 + 配额配置 + 使用报表      | T3.2.2 | console/    | 16h | P0  |

#### Sprint 3.4 — 依赖解析器 v2 + K8s Operator

| 任务 ID  | 任务                                                                                            | 依赖     | 产出             | 预估  | 优先级 |
| :----- | :-------------------------------------------------------------------------------------------- | :----- | :------------- | :-- | :-- |
| T3.4.1 | 依赖解析器 v2：升级为 SAT 求解器（或简化版回溯搜索），处理多版本冲突 + 可选依赖 + 平台条件                                          | T1.1.5 | resolver.go v2 | 16h | P1  |
| T3.4.2 | K8s Operator：`services/operator/`，CRD `PluginInstance` + controller-runtime，管理远程插件的调度/伸缩/滚动升级 | 无      | operator/      | 16h | P1  |

**M3 完成标志**：3 个租户同时使用，租户间数据/插件 100% 隔离；市场上已有 20+ 官方示例插件；插件包推送/安装/下架/回滚全流程可用。

***

### 3.4 M4 · 生产级 + 生态孵化（9\~12 个月）

| 任务 ID | 任务                                                                                             | 依赖        | 产出              | 预估  | 优先级 |
| :---- | :--------------------------------------------------------------------------------------------- | :-------- | :-------------- | :-- | :-- |
| T4.1  | 熔断器：`src/go/host/circuit/`，连续失败计数 → 熔断 → 半开探测 → 恢复                                             | T1.4.4    | circuit/        | 8h  | P0  |
| T4.2  | 故障注入：`src/go/host/chaos/`，测试模式下的延迟/错误/崩溃注入                                                     | 无         | chaos/          | 8h  | P1  |
| T4.3  | Wasm 沙箱加载器：`src/go/host/loader/wasm.go`，嵌入 Wasmtime + WIT 解析 + Component 加载（先用 wasmtime-go 绑定） | 无         | wasm.go         | 24h | P0  |
| T4.4  | Native 动态库加载器：`src/go/host/loader/native.go`，dlopen + C ABI 符号解析（纯 Go 用 `plugin` 标准库，跨语言用 CGO） | 无         | native.go       | 12h | P1  |
| T4.5  | WIT 绑定生成工具：`tools/wit-bindgen/`，从 WIT 接口定义自动生成 6 语言的 Plugin PDK 骨架                             | 无         | wit-bindgen/    | 24h | P1  |
| T4.6  | Tauri 集成模板：`apps/templates/tauri-originlang/`，Tauri 桌面应用嵌入 OriginLang Host SDK                 | T2.1.2    | tauri-template/ | 12h | P2  |
| T4.7  | Theia 集成模板：`apps/templates/theia-originlang/`，Theia IDE 扩展调用 OriginLang 插件系统                   | 无         | theia-template/ | 16h | P2  |
| T4.8  | 性能基准白皮书：`apps/docs/performance-benchmark.md`，4 种载体的延迟/吞吐量/内存基准                                 | T4.3 T4.4 | benchmark       | 8h  | P2  |
| T4.9  | Vue 适配器：`src/ts/core/ui/vue-adapter.ts`                                                        | T2.4.3    | vue-adapter.ts  | 4h  | P2  |
| T4.10 | 事件总线异步后端 (NATS)：`src/go/host/eventbus/async.go`                                                | T1.3.x    | async.go        | 12h | P1  |

***

## 4. 依赖关系图（M1 关键路径）

```
T1.1.1 (错误码)
  ├── T1.4.1 (Security Engine) ──┐
  │       └── T1.4.3 (Manager 集成) ──┐
  └── T1.4.5 (Quota v1) ──────────────┤
                                      └── T1.5.6 (TODO-SaaS)
T1.1.2 (Manifest)
  ├── T1.1.4 (Manifest 测试)
  ├── T1.1.5 (Resolver v1)
  ├── T1.3.3 (Manifest→扩展点注册) ──┐
  │   └── T1.3.1 (ExtensionPoint)   │
  │       └── T1.3.2 (6 内置扩展点)  │
  │           └── T1.3.4 (Query API) ─┤
  └── T3.1.1 (OCI, M3)               │
                                      └── T1.5.6 (TODO-SaaS)
T1.2.1 (Inproc Transport) ──────────┐
T1.2.3 (Kind 枚举扩展) ──┤           │
T1.2.4 (传输选择器)       │           │
T1.2.2 (HTTP Transport, P1)        │
                                      └── T1.5.6 (TODO-SaaS)
T1.5.1 (CLI 骨架)
  ├── T1.5.2 (ol init)
  ├── T1.5.3 (ol build)
  ├── T1.5.4 (ol test)
  └── T1.5.5 (ol run) ──────────────┐
                                      └── T1.5.6 (TODO-SaaS)
T1.6.1 (TS Host SDK)
T1.6.2 (TS Plugin PDK)
T1.6.3 (TS 插件示例) ────────────────┐
T1.6.4 (E2E 测试)                    └── M1 完成
T1.6.5 (Bazel TS)  [并行]
T1.6.6 (Contract Test) [并行]
```

**关键路径**：T1.1.1 → T1.1.2 → T1.3.1 → T1.3.3 → T1.4.1 → T1.4.3 → T1.5.1 → T1.5.6 → T1.6.4

***

## 5. 人力估算汇总

| 里程碑         | 任务数    | 总预估工时      | 1 人月 (140h) | 建议人数                        |
| :---------- | :----- | :--------- | :---------- | :-------------------------- |
| M1 (0-3 月)  | 30     | \~200h     | 1.4 人月      | 1-2 人                       |
| M2 (3-6 月)  | 24     | \~250h     | 1.8 人月      | 2 人                         |
| M3 (6-9 月)  | 10     | \~120h     | 0.9 人月      | 1-2 人                       |
| M4 (9-12 月) | 10     | \~130h     | 0.9 人月      | 1 人                         |
| **合计**      | **74** | **\~700h** | **5.0 人月**  | **2-3 人并行可在 6 个月内完成 M1-M2** |

> 注：以上为净编码工时，不含设计评审、文档、Code Review、调试时间。实际项目周期建议乘以 2-3 倍系数。

***

## 6. 建议的即时启动任务（Top 10）

按优先级排序，建议**立即启动**的前 10 个任务（均可独立开始，无外部依赖）：

| #  | 任务                             | 产出                 | 预估   | 现有代码入口                                                                            |
| :- | :----------------------------- | :----------------- | :--- | :-------------------------------------------------------------------------------- |
| 1  | T1.1.1 扩展错误码                   | error.go +6 常量     | 0.5h | [error.go](file:///e:/code/originlang/src/go/host/rpc/error.go)                   |
| 2  | T1.1.2 Manifest 结构体            | manifest/ 包        | 6h   | 新建                                                                                |
| 3  | T1.2.1 Inproc Transport        | inproc.go          | 4h   | [transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go)     |
| 4  | T1.2.3 Kind 枚举扩展               | transport.go +3 行  | 0.5h | [transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go)     |
| 5  | T1.3.1 ExtensionPoint Registry | extpoint.go        | 4h   | 新建                                                                                |
| 6  | T1.4.1 Security Engine v1      | security/engine.go | 8h   | [manager.go](file:///e:/code/originlang/src/go/host/manager/manager.go)           |
| 7  | T1.4.4 健康检查                    | health/checker.go  | 4h   | [plugin.go](file:///e:/code/originlang/src/go/host/plugin/plugin.go) 的 MethodPing |
| 8  | T1.5.1 CLI 骨架                  | tools/cli/main.go  | 2h   | 新建                                                                                |
| 9  | T1.6.1 TS Host SDK             | src/ts/core/host/  | 12h  | 新建                                                                                |
| 10 | T1.6.2 TS Plugin PDK           | src/ts/core/pdk/   | 6h   | 新建                                                                                |

> 这 10 个任务总计 \~47h，可在 1 周内由 1-2 人完成，作为 M1 Sprint 1 的启动批次。

***

## 7. 现有代码改造清单（需要修改的已有文件）

| 文件                                                                                      | 改造内容                                                                                            | 对应任务                 |
| :-------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------- | :------------------- |
| [rpc/error.go](file:///e:/code/originlang/src/go/host/rpc/error.go)                     | 新增 6 个错误码常量                                                                                     | T1.1.1               |
| [transport/transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go) | Kind 枚举新增 `KindWasm` `KindNative` `KindInproc`                                                  | T1.2.3               |
| [plugin/plugin.go](file:///e:/code/originlang/src/go/host/plugin/plugin.go)             | State 新增 `StateUnloading` `StateDraining`；`Plugin` 增加实例池引用字段                                    | T1.3.5               |
| [manager/manager.go](file:///e:/code/originlang/src/go/host/manager/manager.go)         | `Load()` 方法扩展为多加载器分发（根据 Manifest.artifacts 选择 Transport）；`Call()` 插入 Security Engine + Quota 拦截 | T1.3.3 T1.4.3 T1.4.6 |
| [manager/errors.go](file:///e:/code/originlang/src/go/host/manager/errors.go)           | `toRPCError` 扩展新的错误码映射                                                                          | T1.4.3               |
| [sdk/sdk.go](file:///e:/code/originlang/src/go/host/sdk/sdk.go)                         | `Serve()` 增加从 manifest.json 读取 Name/Version/Capabilities 的能力                                    | T1.5.2               |
| [MODULE.bazel](file:///e:/code/originlang/MODULE.bazel)                                 | 取消注释 TS 规则依赖                                                                                    | T1.6.5               |
| [src/go/BUILD.bazel](file:///e:/code/originlang/src/go/BUILD.bazel)                     | 新增 manifest/extpoint/security/quota/observ/health/hotreload 子包依赖                                | M1 全程                |
| [src/BUILD](file:///e:/code/originlang/src/BUILD)                                       | 新增 ts/core 的 filegroup 引用                                                                       | T1.6.1               |

***

> **下一步行动**：建议从 §6 的 Top 10 任务开始，先完成 T1.1.1（0.5h，零风险改常量）和 T1.1.2（6h，新建 manifest 包），作为 Sprint 1.1 的第一周任务。同步启动 T1.6.1/T1.6.2 的 TS SDK/PDK（与 Go 内核改造无依赖，可并行）。

