# OriginLang 平台底座 · 架构设计草案

| 版本         | 日期         | 作者             | 状态     | 说明                                              |
| :--------- | :--------- | :------------- | :----- | :---------------------------------------------- |
| v0.1-Draft | 2026-09-02 | OriginLang 架构组 | **草案** | 初版架构设计，基于 GitHub 同类项目调研结论 + 现有 `src/go/host` 实现 |

***

## 1. 项目愿景与设计目标

### 1.1 愿景

**OriginLang = 一个"插件化、支持多编程语言"的通用平台应用底座。**

任何团队/个人，只要基于 OriginLang 构建产品（SaaS 平台、开发者工具、AI 工作台、企业中台、桌面/移动应用……），都能获得：

1. **多语言写插件**：插件开发者可用 **Go / Rust / Java / Python / C++ / TypeScript** 六大主流语言任选其一写后端逻辑插件，用 **TS/JS + 主流前端框架 (React/Vue)** 写 UI 扩展；
2. **多语言写宿主**：平台宿主程序同样可选上述六大语言嵌入 OriginLang Kernel；
3. **动态插件市场级能力**：运行时热加载/热更新/热卸载、版本依赖解析、细粒度权限、插件市场分发；
4. **SaaS 原生多租户**：租户级插件可见性、配额、计费计量、故障隔离；
5. **一体前后端扩展**：一个插件包同时声明"后端 API + 前端组件/菜单/路由"，宿主零代码注入。

### 1.2 非目标

- **不做** 单一垂直领域产品（如 IDE、AI Agent 平台、CRM）；OriginLang 是**底座**，不是成品；

- **不做** 自己的编程语言；"Lang" 指多语言插件，不是新语言；

- **不重复造** 成熟基础轮子：传输协议用 JSON-RPC 2.0（已实现）、Wasm 运行时用 Wasmtime、容器编排用 K8s Operator、构建系统用 Bazel。

### 1.3 设计原则（针对竞品不足的规避）

| #  | 原则                                                         | 对应竞品问题的针对性                                                  |
| :- | :--------------------------------------------------------- | :---------------------------------------------------------- |
| P1 | **语言无关优先，宿主 SDK 与插件 PDK 六语言齐平发布**                          | Extism 长尾 SDK 冻结 / Tauri 必须写 Rust / Theia 仅 TS / PF4J 只 JVM |
| P2 | **协议与传输解耦：一种 JSON-RPC 协议跑在 stdio/TCP/HTTP/Wasm-IPC 四种传输上** | waPC 社区停滞 / 多数项目只有一种部署形态                                    |
| P3 | **四种插件载体共存、统一生命周期**                                        | Extism/Wasm 不做原生高性能、PF4J 只 JVM、Tauri 编译期绑定                  |
| P4 | **前后端扩展一体设计：一套 Manifest 同时声明后端能力 + UI 贡献点**                | 几乎所有竞品都分属两套体系                                               |
| P5 | **Deny-by-Default 安全 + 细粒度 Capabilities + 租户级二次授权**        | Tauri 插件同进程无隔离 / Extism 无租户概念                               |
| P6 | **多租户一等公民：插件仓库可见性、实例隔离、配额计量**                              | 所有竞品都无内建多租户                                                 |
| P7 | **可观测性一体化：插件级指标/日志/追踪零配置接入**                               | Wasm 类项目调试难、Theia 慢扩展难定位                                    |

***

## 2. 总体分层架构

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  L8 · 应用集成层 (Application Integration Layer)                                     │
│  领域产品 / 企业中台 / AI 工作台 / SaaS 控制台（OriginLang 不提供，只给接口）          │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L7 · 平台治理层 (Platform Governance)                                               │
│  插件市场 · OCI 注册中心 · 版本/签名/升级/回滚 · 多租户控制台 · 运营分析              │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L6 · 前端 UI 扩展层 (UI Extension Layer)                                            │
│  Web Components + Module Federation · 贡献点: 菜单/路由/视图/组件/设置/状态栏        │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L5 · 插件运行时层 (Plugin Runtime Layer)                                            │
│  ① Wasm 沙箱运行时 (Wasmtime)  ② 子进程运行时  ③ 原生 .so/.dll  ④ 远程 HTTP/TCP    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L4 · 多语言 SDK / PDK 层 (Polyglot SDK Layer)                                      │
│  Host SDK:     Go | Rust | Java | Python | C++ | TypeScript （6 种一等公民）        │
│  Plugin PDK:   Go | Rust | Java | Python | C++ | TypeScript （6 种一等公民）        │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L3 · 协议与传输层 (Protocol & Transport Layer)                                      │
│  JSON-RPC 2.0 (已实现) · Manifest Schema (WIT/JSON) · Codec · stdio/TCP/HTTP/Wasm   │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L2 · OriginLang 内核层 (Kernel Layer)  —  唯一核心，用 Go + Rust 双实现共享 ABI     │
│  ┌──────────┬─────────────┬──────────────┬──────────┬──────────┬──────────────────┐ │
│  │ Plugin   │ Extension   │ Event Bus    │ Security │ Resource │ Observability    │ │
│  │ Manager  │ Point Reg.  │ (Pub/Sub)    │ Engine   │ Quota    │ Collector        │ │
│  ├──────────┼─────────────┼──────────────┼──────────┼──────────┼──────────────────┤ │
│  │ Lifecycle│ Dependency  │ Service      │ Tenant   │ Hotload  │ Health Check     │ │
│  │ StateMach│ Resolver    │ Discovery    │ Isolation│ Manager  │ + Circuit Breaker│ │
│  └──────────┴─────────────┴──────────────┴──────────┴──────────┴──────────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L1 · 基础设施层 (Infrastructure Layer)                                              │
│  Bazel 多语言构建 · Wasmtime 引擎 · K8s Operator · OCI Registry · PostgreSQL/Redis  │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.1 内核实现策略

- **内核主体 (Kernel Core)**：**Go 语言实现**（现仓库 `src/go/host/` 已有的 rpc/transport/plugin 三个模块作为起点），理由：

  - Go 在服务端并发、网络、可部署性上最佳；

  - 现仓库已打下约 1000 行的成熟骨架（State 状态机、Endpoint、JSON-RPC 2.0 解包、Capability 结构）。

- **性能敏感子模块（Wasm 引擎桥接、资源配额 Enforcer）**：**Rust 实现**，通过 C ABI（FFI）供 Go 内核调用；

- **宿主多语言适配**：Go/Rust 内核导出 C ABI → 各语言 Host SDK 做薄绑定（类似 Extism 的 `libextism` 策略，但我们的 `liboriginlang` 同时导出 Go 核心 + Rust 高性能模块）。

***

## 3. 各层详细设计

### 3.1 L1 基础设施层

| 模块          | 选型                                       | 职责                                       |
| :---------- | :--------------------------------------- | :--------------------------------------- |
| **构建系统**    | **Bazel 9 (MODULE.bazel)**（已在仓库中）        | 六语言统一构建、跨平台产物、缓存、Hermetic 测试             |
| **Wasm 引擎** | **Wasmtime 43+** + Component Model + WIT | Wasm 沙箱插件运行时（对标 Extism/wasmCloud）        |
| **容器编排**    | **K8s + OriginLang Operator**            | 分布式部署下的插件调度、伸缩、滚动升级                      |
| **插件分发**    | **OCI Registry (ORAS 协议)**               | 插件包 = OCI Artifact；支持签名（Sigstore/cosign） |
| **元数据存储**   | PostgreSQL                               | 插件元数据、版本、租户授权、配额、审计日志                    |
| **缓存 / 事件** | Redis + 可选 NATS                          | 事件总线后端、插件实例池缓存、分布式锁                      |

***

### 3.2 L2 内核层（核心 6 大子系统）

基于现仓库 `src/go/host/plugin.Plugin` 扩展：

#### 3.2.1 插件管理器 (Plugin Manager) — 现有基础最完善

现有 `State` 状态机已定义：

```
Created → Starting → Registered → Ready → Running → Stopped
                                          ↘ Failed
```

**新增：**

- **加载器多态化**：根据插件载体类型调用 4 种运行时加载器之一 (见 3.5)

- **实例池**：Wasm/原生插件支持 `pool_size` 配置（类似数据库连接池），避免每次请求实例化；子进程/远程插件支持连接池

- **健康检查**：`plugin.ping`（已定义常量）周期心跳；连续失败自动 Failover 到备用实例

#### 3.2.2 扩展点注册表 (Extension Point Registry) — 全新（核心差异化）

OriginLang 区别于"纯 RPC 插件系统"的关键：引入 **Contribution Point (扩展点)** 模型。

```go
// 伪代码：内核中扩展点与扩展的抽象
type ExtensionPoint struct {
    ID          string         // 例如 "originlang.ui.sidebar.menu"
    Description string
    Schema      JSONSchema/WIT // 扩展必须实现的接口契约
    Scope       Scope          // Global / Tenant / User
}

type Extension struct {
    PluginID    string         // 来自哪个插件
    PointID     string         // 挂到哪个扩展点
    Order       int            // 排序权重
    Enabled     bool
    Config      map[string]any // 扩展级配置
}
```

**内置扩展点（平台自带，不依赖业务）：**

| 扩展点 ID                        | 用途        | 说明                                                        |
| :---------------------------- | :-------- | :-------------------------------------------------------- |
| `originlang.api.handler`      | 后端 API 处理 | 插件声明 HTTP/gRPC/JSON-RPC 路由；宿主自动注册到网关                      |
| `originlang.pipeline.hook`    | 通用流水线钩子   | 如 `before_request` / `after_response` / `event_transform` |
| `originlang.data.source`      | 数据源连接器    | JDBC/Redis/ES/Mongo 等连接提供者                                |
| `originlang.auth.provider`    | 认证提供者     | OIDC/LDAP/SAML/OAuth2 等                                   |
| `originlang.job.scheduler`    | 任务调度扩展    | Cron / DAG 工作流节点类型                                        |
| `originlang.ui.menu`          | 前端菜单注入    | 类型=菜单                                                     |
| `originlang.ui.route`         | 前端路由注入    | 懒加载 MF 模块                                                 |
| `originlang.ui.component`     | 前端组件注入    | Web Components / React 组件                                 |
| `originlang.ui.setting_panel` | 前端设置面板    | 插件专属设置页                                                   |
| `originlang.ui.status_bar`    | 状态栏小部件    | 桌面/IDE 风格状态栏                                              |

#### 3.2.3 事件总线 (Event Bus)

- **两种模式共存**：

  1. **进程内同步 Event**：`sync.PubSub`，用于单实例部署下插件间轻量通信；
  2. **分布式异步 Event (NATS / Redis Stream)**：`async.PubSub`，subject 命名规范：`originlang.{tenant}.{domain}.{event}`；

- 支持**通配符订阅**和**事件模式过滤**（CloudEvents 格式 1.0）。

#### 3.2.4 安全引擎 (Security Engine) — 详细设计见 §6

- 三层权限：`Plugin Capability Manifest` → `Tenant Grant` → `Runtime User Policy`

- 运行时拦截器：内核对每个 `plugin.call` 执行：

  1. 方法是否在插件声明 `Capability.Method` 中？
  2. 是否命中当前租户 Grant 的 Allow/Deny 列表？
  3. 是否命中当前用户的 RBAC/ABAC Policy？
  4. 是否触发 Resource Quota（QPS/并发/CPU 时间/内存）？

#### 3.2.5 资源配额 (Resource Quota)

```yaml
# 租户-插件级配额配置示例
tenant_acme.plugins.my_data_connector:
  qps: 1000
  concurrent_calls: 50
  cpu_seconds_per_hour: 3600      # Wasm/子进程 CPU 时间
  memory_mb_per_instance: 512     # 单实例 RSS 硬限
  daily_network_mb: 10240
  timeout_ms_per_call: 5000
```

- **Enforcer 实现**：Rust 写的高性能令牌桶 + 红黑树 CPU 计时（通过 FFI 给 Go 用），每一次 Call 进入/退出时原子更新。

- 超限：返回 `rpc.ErrCodeQuotaExceeded = -32005`，并触发熔断（连续超限 3 次 → 插件 60 秒拒绝所有请求）。

#### 3.2.6 可观测性采集器 (Observability Collector)

- 每一个 `Endpoint.Call` 自动生成：

  - **Metric**（Prometheus）：`originlang_plugin_call_duration_ms{plugin,method,status,tenant}` / `originlang_plugin_call_total{...}` / `originlang_plugin_cpu_seconds{...}`

  - **Trace**（OpenTelemetry）：自动注入 `traceparent` 到 RPC params，跨插件/跨宿主全链路；

  - **Log**（结构化 JSON）：统一字段 `plugin_id / tenant_id / method / req_id / duration_ms / error_code`，自动关联 trace\_id；

- 慢插件检测：`P95 > threshold × 3` 自动输出 CPU Profile（Wasm 用 Wasmtime Profiling API，子进程用 SIGPROF pprof）。

***

### 3.3 L3 协议与传输层

**协议沿用现有设计 (JSON-RPC 2.0)**，只做增量扩展：

1. **标准错误码扩展**（现有 `rpc/error.go` 基础上）：

| 错误码    | 常量                         | 说明                    |
| :----- | :------------------------- | :-------------------- |
| -32001 | `ErrCodePluginStopped`     | 已定义（endpoint closed）  |
| -32002 | `ErrCodeUnauthorized`      | 权限拒绝（Security Engine） |
| -32003 | `ErrCodeMethodForbidden`   | 方法未在 Capability 中声明   |
| -32004 | `ErrCodeTenantIsolated`    | 跨租户访问拒绝               |
| -32005 | `ErrCodeQuotaExceeded`     | 资源配额超限                |
| -32006 | `ErrCodeCircuitBroken`     | 熔断开启                  |
| -32007 | `ErrCodeDependencyMissing` | 插件依赖不满足               |

1. **Manifest 规范**（每个插件包根目录必须包含 `originlang-manifest.json`，版本化 Schema）：

```json
{
  "$schema": "https://originlang.dev/schemas/manifest-v1.json",
  "id": "com.acme.data-connector",
  "name": "Acme Data Connector",
  "version": "1.4.2",
  "originlang_min_version": "0.3.0",
  "authors": ["team@acme.com"],
  "license": "Apache-2.0",
  "description": "Connects to Acme SaaS data sources",

  // ---- 插件载体 ----
  "artifacts": {
    "wasm":     "dist/plugin-component.wasm",   // ① 首选（Wasm Component Model）
    "binary": {                                 // ② 次选（原生动态库/可执行）
      "linux-amd64":   "dist/plugin-linux-amd64.so",
      "darwin-arm64":  "dist/plugin-mac-arm64.dylib",
      "windows-amd64":"dist/plugin-win-amd64.dll"
    },
    "remote": {                                 // ③ 远程 HTTP 服务地址（可选）
      "default_endpoint": "https://plugin.acme.com/rpc"
    }
  },

  // ---- 依赖管理 ----
  "dependencies": [
    { "id": "originlang.stdlib", "version": ">=0.2.0" },
    { "id": "com.example.db-pool", "version": "~2.1.0", "optional": true }
  ],

  // ---- 能力声明（Deny-by-Default，这里声明"我需要什么权限"） ----
  "capabilities": {
    "rpc_methods": [                 // 我（作为 Server）对外暴露哪些 JSON-RPC 方法
      { "method": "acme.query",      "description": "Query Acme SaaS data",
        "timeout_ms": 3000, "memory_mb": 128 },
      { "method": "acme.export",     "description": "Export data to CSV",
        "timeout_ms": 30000, "memory_mb": 512 }
    ],
    "host_calls": [                  // 我（作为 Client）需要调用宿主提供的哪些能力
      "originlang.kv.get", "originlang.kv.set",         // KV 存储
      "originlang.http.fetch",                            // 外呼 HTTP
      "originlang.fs.read:./data/**",                    // 文件系统（路径范围）
      "originlang.events.publish:acme.data.*"            // 事件发布（subject 模式）
    ],
    "env_vars": ["ACME_API_KEY"],    // 需要读取哪些环境变量
    "network_egress": [              // 需要外连的网络
      "https://*.acme.com/*",
      "tcp://mq.acme.com:5671"
    ]
  },

  // ---- 扩展点贡献 ----
  "extensions": [
    { "point": "originlang.data.source",
      "name": "acme_source",
      "order": 100,
      "config_schema": { "$ref": "./config-schema.json" } },
    { "point": "originlang.ui.menu",
      "props": { "label": "Acme Data", "icon": "database", "route": "/acme" } },
    { "point": "originlang.ui.route",
      "props": { "path": "/acme", "mf_module": "./dist/ui/acme-remote.js" } }
  ]
}
```

1. **传输矩阵（对现有** **`transport.Kind`** **扩展三种）**：

| Kind          | 场景                     | 隔离级别         | 典型延迟    |
| :------------ | :--------------------- | :----------- | :------ |
| `stdio` (已实现) | 本地子进程插件（任何可执行文件）       | OS 进程级       | \~100μs |
| `tcp` (已实现)   | 本地或远端 TCP 长连接          | 网络级          | \~200μs |
| `http` (规划)   | 远程微服务插件（部署在 K8s 上）     | 网络级          | \~1ms   |
| `wasm` (新增)   | 进程内 Wasm 沙箱插件          | Wasmtime 沙箱级 | \~1μs   |
| `native` (新增) | 进程内原生动态库插件 (dlopen)    | 同进程（弱隔离）     | \~10ns  |
| `inproc` (新增) | 同语言同进程直接调用（宿主与插件语言一致时） | 同进程（无隔离）     | <1ns    |

**传输选择算法**（内核按优先级自动选择，Manifest 提供多种 artifact 时）：

```
inproc (同语言)  →  native (有匹配平台 .so/.dll)  →  wasm (首选，跨语言+安全)
           ↘ 不可用或配置禁用       ↘ 不支持原生 ABI ↗
                                            stdio (通用，任何可执行)  →  remote (HTTP)
```

***

### 3.4 L4 多语言 SDK / PDK 层（六大语言齐平）

这是 OriginLang 最核心的差异化（针对 Extism 长尾 SDK 冻结）。**发布策略：任何 Host SDK / Plugin PDK 的大版本必须 6 语言同时发布，否则不发版。**

\| 语言 | Host SDK 定位 | Plugin PDK 定位 | 底层依赖 |
\| :-- | :-- | :-- |
\| **Go** | 宿主内核参考实现 (本仓库) | `tinygo build -target wasip1` + Go 子进程 | 纯 stdlib 依赖（现有代码已经是零第三方依赖） |
\| **Rust** | 性能敏感宿主；Wasm 桥接层 | `cargo build --target wasm32-wasip1` + 原生 cdylib | `tokio` (异步) + `wasmtime` (可选嵌入) |
\| **Java** | 企业级宿主（Spring Boot 集成） | `GraalVM Native Image → wasm` + Java 子进程 JVM | `Foreign Function & Memory API` (Java 22+) 调用 `liboriginlang` |
\| **Python** | AI / 数据科学宿主 | `Pyodide wasm` + Python 子进程 CPython | `cffi` 绑定 `liboriginlang` |
\| **C++** | 高性能 / 嵌入式宿主 | 原生动态库直接 dlopen + Emscripten wasm | C ABI 头文件（`originlang.h`） |
\| **TypeScript** | Web / Node.js 宿主 + 所有 UI 插件作者 | `JCO` (Component Model → JS) + Node 子进程 | `node:ffi-napi` 或纯 JS 重实现（像 Extism JS SDK 那样） |

***

### 3.5 L5 插件运行时层（四种载体共存，统一生命周期）

| 载体                            | 实现方式                                                                      | 优点                                              | 缺点                                                | 适用场景                            |
| :---------------------------- | :------------------------------------------------------------------------ | :---------------------------------------------- | :------------------------------------------------ | :------------------------------ |
| **① Wasm 沙箱（默认推荐）**           | Wasmtime 加载 Wasm Component；`wasm` 传输 = 共享内存 IPC；接口用 WIT 定义 + JSON-RPC 2.0 | **最强隔离**、跨平台跨语言、最小体积 (1-5MB)、Deny-by-Default 原生 | 性能比原生低 5-30%；GC 语言 (Java/Py) 的 Wasm 包体积大 (8-15MB) | 第三方不可信插件、SaaS 多租户市场、AI 生成代码运行   |
| **② 子进程（已实现 stdio/tcp）**      | fork/exec 插件可执行文件；stdio JSON-RPC 或 TCP localhost 连接                       | 任何可执行语言都能写；完全进程级隔离；崩溃不影响宿主                      | 启动慢 (10-1000ms)；IPC 有开销；内存占用高                     | 现有遗留系统快速接入；重型计算插件；Python 数据科学生态 |
| **③ 原生动态库 (.so/.dll/.dylib)** | `dlopen` / `LoadLibrary` 加载 C ABI 符号；`native` 传输                          | **最低延迟** (ns 级)；零拷贝内存共享                         | 同进程无隔离，一个段错误宿主全崩；跨平台编译复杂                          | 自写的、完全可信的、高吞吐量插件（内部网关/加密/序列化）   |
| **④ 远程服务 (HTTP/TCP)**         | 插件跑在远端服务器/K8s Pod 上；`http`/`tcp` 传输                                       | 插件独立扩缩容；可跑 GPU/数据库等宿主不具备的资源                     | 网络延迟；运维成本；SLA 绑定                                  | 微服务化部署；GPU 推理插件；跨团队独立发布的团队插件    |

***

### 3.6 L6 前端 UI 扩展层（核心差异化，竞品普遍缺失）

**设计目标：后端插件作者用任何语言写插件时，都可以"顺手"附带一个前端 UI 扩展包，宿主应用自动渲染，无需改代码。**

#### 3.6.1 前端贡献机制

基于 **Web Components (Custom Elements v1 + Shadow DOM)** 作为 UI 组件的**跨框架标准载体**，叠加 **Module Federation (MF)** 用于懒加载大体积业务模块。

```
插件 UI 包结构：
dist/ui/
├── acme-remote.js          ← Module Federation remoteEntry（React/Vue 业务模块）
├── acme-menu-item.js       ← Web Component：<acme-menu-item>
├── acme-dashboard.js       ← Web Component：<acme-dashboard>
└── acme-settings.js        ← Web Component：<acme-settings-panel>
```

宿主应用（框架无关，React/Vue/Angular/Svelte 都可以）只需：

1. 启动时从内核拉取 `extensions["originlang.ui.*"]` 列表；
2. 对 `ui.menu` 扩展：注册菜单树，点击后导航到对应路由；
3. 对 `ui.route` 扩展：路由匹配时懒加载 MF 模块，或直接插入对应的 Web Component。
4. 所有组件通过标准 Custom Events + 统一的 `window.OriginLangHost` 对象（暴露 `host.call(method, params)`）与后端通信，**不依赖任何前端框架**。

#### 3.6.2 UI 权限与可见性

- 菜单/路由/组件**级别的权限**：插件 Manifest 中每个 UI 扩展可声明 `required_permissions: ["acme.view", "acme.admin"]`，宿主拉取扩展时按当前用户权限过滤。

- 多租户**级别的可见性**：租户未授权该插件 → 完全不返回任何 UI 扩展。

***

### 3.7 L7 平台治理层

| 子模块          | 功能                                                                                                |
| :----------- | :------------------------------------------------------------------------------------------------ |
| **OCI 注册中心** | 插件包 = OCI Artifact，支持 `oras push/pull`；支持 cosign 签名；支持 semver tag + channel (stable/beta/nightly) |
| **插件市场 Web** | 浏览/搜索/评级/评论/使用统计；商业插件可对接支付；安全扫描报告 (Wasm security audit + SBOM)                                    |
| **版本与升级**    | 依赖冲突自动解析（Maven/npm 同款 SAT 求解器）；灰度升级（按租户 1%→10%→50%→100%）；一键回滚                                     |
| **多租户控制台**   | 租户管理员：授权/禁用插件、配置配额、查看账单与使用报表；平台管理员：全局插件下架/熔断                                                      |
| **运营分析**     | 插件安装量/活跃租户/DAU/错误率/性能 P95 大盘；异常插件自动告警给作者                                                          |

***

### 3.8 L8 应用集成层

OriginLang 本身**不做任何垂直业务功能**。这一层留给业务方，典型集成方式：

- **企业中台团队**：基于 OriginLang Host SDK（Java/Spring Boot）搭中台，业务部门各自写 Go/Java/Python 插件 + UI 扩展；

- **AI 工作台产品**：基于 OriginLang Host SDK（Python + TS）搭平台，各模型供应商、各工具链都是 OriginLang 插件；

- **开发者工具（IDE/Cloud IDE）**：基于 OriginLang Host SDK（Go/Rust + TS）搭核心，语言服务、调试器、Linter 都是插件；

- **跨平台桌面应用**：基于 OriginLang Host SDK（Rust）嵌入 Tauri，OriginLang 管"业务插件系统"，Tauri 管"窗口/系统 API"。

***

## 4. 关键流程设计

### 4.1 插件安装（租户管理员操作）

```
Admin ──授权──▶ 平台治理层
   │                │
   │                1. 从 OCI 拉取插件包，验证签名 + SBOM
   │                2. 解包解析 Manifest → 校验 dependencies 版本兼容
   │                3. 写入 plugin_versions 表 + tenant_plugin_grants 表
   │                4. 给插件分配资源配额（默认模板 / 自定义）
   │                5. 通知内核插件管理器"插件可用"
   │
   ▼
  内核开始 惰性加载（首次调用才实例化）
        ↓
   选择载体 (inproc → native → wasm → stdio → remote)
        ↓
   Created → Starting → (handshake plugin.register) → Registered → Ready → Running
```

### 4.2 插件调用（用户在租户下发起请求）

```
User Request → 宿主 API Gateway
        │
        ▼
  OriginLang 内核 Security Engine
   ├─① 方法在 Capability 中声明？→ 否则 ErrCodeMethodForbidden
   ├─② 租户 Grant 允许？→ 否则 ErrCodeTenantIsolated
   ├─③ 用户 RBAC 允许？→ 否则 ErrCodeUnauthorized
   ├─④ 配额 (QPS/并发) 未满？→ 否则 ErrCodeQuotaExceeded
        ▼
  Resource Quota Enforcer（Rust FFI，原子计数）
        ▼
  插件实例池取实例 → Endpoint.Call(method, params)
        │
        ├── wasm/native → 进程内调用 (ns~μs)
        ├── stdio/tcp   → JSON-RPC 2.0 写 pipe/socket (μs 级)
        └── http/remote → HTTPS JSON-RPC 2.0 (ms 级)
        ▼
  结果返回（成功/错误 → 自动写指标/追踪/日志）
        │
        └──错误率/超时率高？→ 自动熔断 → ErrCodeCircuitBroken
```

### 4.3 热升级（不中断用户请求）

```
1. 安装新版本插件包 → 内核标记 v1.4.2 "待切换"，当前运行实例仍是 v1.4.1
2. 预热：启动 N 个新实例 → Running，但还不接受流量
3. 流量切分：按百分比（1%/10%/50%/100%）把请求路由到新版本
4. 观测：新版本 error_rate / p95 若连续 5 分钟超阈值 → 自动回滚
5. 老版本实例：等所有 in-flight 请求结束后，逐一 Shutdown 回收
```

***

## 5. 与现有仓库的映射关系

```
originlang/
├── src/
│   ├── go/
│   │   ├── host/                          ← 【L2 内核 · Go 参考实现】
│   │   │   ├── rpc/        (已有, JSON-RPC 2.0)  → 扩展 §3.3 错误码 + 类型安全 Wrapper
│   │   │   ├── transport/  (已有, stdio/tcp)     → + http/wasm/native 三种 Transport
│   │   │   ├── plugin/     (已有, StateMachine)  → + 实例池 + 4 种加载器 + 健康检查
│   │   │   ├── kernel/     (新增)                 → 集成 6 大子系统的 Facade
│   │   │   ├── security/   (新增)                 → Security Engine + 三层权限
│   │   │   ├── quota/      (新增)                 → 绑定 Rust libquota_enforcer.so FFI
│   │   │   ├── extpoint/   (新增)                 → ExtensionPoint Registry + 10 个内置点
│   │   │   ├── eventbus/   (新增)                 → sync + async (NATS/Redis) 双后端
│   │   │   ├── observ/     (新增)                 → OpenTelemetry 自动埋点
│   │   │   └── pkg/        (新增)                 → OCI 包下载/签名校验/依赖解析
│   │   └── sdk/                                  ← 【L4 · Go Host SDK】（包一层 kernel Facade + 类型安全）
│   │   └── pdk/                                  ← 【L4 · Go Plugin PDK】（子进程插件骨架 + Wasm TinyGo 模板）
│   │
│   ├── rust/
│   │   ├── core/         (新增)                   ← 【L2 · 高性能内核子模块 + liboriginlang.so C ABI】
│   │   │   ├── quota_enforcer/   → 资源配额令牌桶 + CPU 时间计时
│   │   │   ├── wasm_bridge/      → Wasmtime 嵌入 + WIT 绑定生成
│   │   │   ├── c_abi/            → 导出 originlang_*() C 函数供 6 语言 SDK 绑定
│   │   │   └── fuzz_tests/       → 安全模糊测试
│   │   ├── sdk/          (新增)                   ← 【L4 · Rust Host SDK】
│   │   └── pdk/          (新增)                   ← 【L4 · Rust Plugin PDK】
│   │
│   ├── java/core/      (新增)                     ← 【L4 · Java Host SDK + Plugin PDK】FFM 调用 liboriginlang
│   ├── python/core/    (新增)                     ← 【L4 · Python Host SDK + Plugin PDK】cffi 调用 liboriginlang
│   ├── cpp/core/       (新增)                     ← 【L4 · C++ Host SDK + Plugin PDK】头文件 + 链接 liboriginlang
│   └── ts/core/        (新增)                     ← 【L4 · TS Host SDK + Plugin PDK + UI SDK（Web Components + MF）】
│
├── apps/
│   ├── docs/                                  ← 架构文档（本文件）
│   ├── market-web/   (未来 M3)                ← 【L7 · 插件市场 Web 应用】
│   └── console/      (未来 M3)                ← 【L7 · 多租户管理控制台】
│
├── services/
│   ├── registry/     (未来 M3)                ← 【L7 · 插件 OCI Registry 适配层 + 元数据 DB】
│   └── otel-collector-config/ (未来 M2)       ← 【L2 · 可观测性 OpenTelemetry Collector 配置】
│
├── tools/
│   ├── bazel_rules/  (新增)                   ← OriginLang 自定义 Bazel 规则：
│   │   ├── originlang_plugin.bzl              →   originlang_plugin(manifest, wasm_srcs, ui_srcs, ...)
│   │   └── originlang_sdk_dep.bzl             →   统一依赖六语言 SDK 版本
│   └── cli/          (新增)                   ← `ol` 命令行工具（对标 wasmCloud wash / Extism xtp）：
│                                                  ol plugin init / build / test / push / install / run / dev
│
├── tests/
│   ├── integration/  (未来 M2)                ← 跨语言 E2E：Go 宿主 ↔ 6 语言插件，矩阵跑 4 种传输
│   └── security/     (未来 M2)                ← 安全模糊测试 + 权限绕过尝试 + 多租户隔离测试
│
└── third_party/                              ← Bazel 第三方依赖
```

***

## 6. 路线图（4 个里程碑）

| 里程碑                  | 时间       | 交付物                                                                                                                                                  | 完成标志                                                                                                |
| :------------------- | :------- | :--------------------------------------------------------------------------------------------------------------------------------------------------- | :-------------------------------------------------------------------------------------------------- |
| **M1 · 内核闭环 (MVP)**  | 0\~3 个月  | Go 内核：完善 4 种传输 + 插件 4 载体加载器 + 扩展点注册表（前 6 个后端扩展点）+ Security Engine v1 + Resource Quota v1；`ol` CLI (init/build/test)；Go/TS Host SDK + Go/TS PDK       | 一个示例应用（TODO SaaS）：3 个插件（Go Wasm / TS stdio / Rust native），1 个菜单 UI 扩展，1 个 API 扩展，单机跑通全生命周期 + 权限拦截生效 |
| **M2 · 六语言齐平 + 可观测** | 3\~6 个月  | Rust/Java/Python/C++ 四种 Host SDK + Plugin PDK 全部 v1.0；可观测性 (Prom/OTel)；`integration/` 跨语言矩阵测试；前端 UI SDK v1（菜单/路由/组件/设置面板 4 个 UI 扩展点）；热升级/回滚          | CI 中 6 语言 × 4 载体的 E2E 测试全绿；P95 延迟、错误率在 Grafana 大盘可见；10 插件同时升级无中断                                    |
| **M3 · 多租户 + 插件市场**  | 6\~9 个月  | OCI Registry + cosign 签名；插件市场 Web + 多租户控制台；依赖冲突 SAT 求解器；灰度升级/降级；事件总线分布式后端 (NATS)；远程 HTTP 插件 + K8s Operator v1                                        | 3 个租户同时使用，租户间数据/插件 100% 隔离；市场上已有 20+ 官方示例插件；插件包推送/安装/下架/回滚全流程可用                                     |
| **M4 · 生产级 + 生态孵化**  | 9\~12 个月 | 熔断/限流/降级/故障注入完整；UI 扩展点完善至 10 个；Wasm Component Model WIT 自动绑定生成工具（对标 XTP Bindgen）；Tauri + OriginLang 集成模板；Theia + OriginLang 集成模板；性能白皮书（4 种载体延迟/吞吐基准） | 至少一个真实业务团队生产使用；社区贡献 ≥3 个第三方 Host SDK 或 PDK；GitHub Stars ≥1k（目标）                                     |

***

## 7. 与同类项目的差异化总结（对竞品不足的直接回应）

| 维度                    | Extism              | wasmCloud              | Tauri v2           | Eclipse Theia    | PF4J             | TEN            | **OriginLang**                                |
| :-------------------- | :------------------ | :--------------------- | :----------------- | :--------------- | :--------------- | :------------- | :-------------------------------------------- |
| **插件可选语言**            | 10+（但长尾冻结）          | 5（Java 早期）             | ❌ 插件必须 Rust        | ❌ 仅 TS/JS        | ❌ 仅 JVM          | 3（C++/Go/Py）   | ✅ **6 语言齐平发布**（Go/Rust/Java/Py/C++/TS）        |
| **宿主可选语言**            | 16+（长尾冻结）           | ❌ 主要 Rust/Go CLI       | ❌ 仅 Rust           | ❌ 仅 Node         | ❌ 仅 Java         | ❌ 主要 C++/Go    | ✅ **6 语言齐平**                                  |
| **前端 UI 扩展**          | ❌ 无                 | ❌ 无                    | ✅（但必须 Rust 后端）     | ✅（仅 TS/IDE 场景）   | ❌ 无              | ❌ 无            | ✅ **Web Components + MF 跨框架一体化**              |
| **四种插件载体并存**          | ❌ 仅 Wasm            | ❌ 仅 Wasm 组件            | ❌ 编译期绑定 Rust crate | ❌ npm 包（编译期）     | ❌ 仅 jar（JVM）     | ❌ 原生模块         | ✅ **Wasm/子进程/原生库/远程** 四种统一生命周期                |
| **运行时热加载**            | ✅                   | ✅                      | ❌（编译期绑定）           | ⚠️（dev 模式可热载）    | ✅                | ✅              | ✅ 且**四种载体都支持热加载/热升级**                         |
| **细粒度 Capability 权限** | ⚠️ 仅 Wasm 级         | ✅ WASI deny-by-default | ⚠️ 仅命令级            | ⚠️ 仅 workspace 级 | ❌ 无内建            | ⚠️ 简单          | ✅ **三层：Manifest + 租户 Grant + 用户 RBAC + 资源配额** |
| **一等多租户能力**           | ❌ 无                 | ❌ 无                    | ❌ 无                | ❌ 无              | ❌ 无              | ❌ 无            | ✅ **内建：可见性/配额/计费/实例隔离**                       |
| **插件级可观测性**           | ❌ 无                 | ⚠️ 基础 tracing          | ❌ 无                | ⚠️ 少             | ❌ 无              | ⚠️ 实时场景强       | ✅ **零配置指标/日志/追踪/慢插件剖析**                       |
| **插件依赖版本管理**          | ❌ 需自建               | ⚠️ 组件组合无 SAT           | ⚠️ Cargo 解决        | ⚠️ npm 地狱        | ⚠️ 简易 + Maven 冲突 | ❌ 无            | ✅ **SAT 求解器 + semver + 灰度 + 回滚**              |
| **部署形态**              | Server/Edge/CLI/IoT | Cloud/Edge K8s         | 桌面/移动 App          | Browser/Electron | JVM Server       | Server/SDK     | ✅ **全部覆盖**：单机/K8s/桌面(Tauri)/移动/边缘/IDE(Theia)  |
| **定位**                | 通用 Wasm 插件系统        | K8s 级 Wasm 微服务平台       | 跨平台应用框架            | IDE/开发工具平台       | Java 服务端模块化      | 实时 AI Agent 框架 | **通用多语言插件化平台底座（所有场景）**                        |

***

## 8. 风险与开放问题（待讨论）

| #  | 风险/问题                                             | 建议                                                                                                                                    |
| :- | :------------------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------ |
| R1 | 六语言 SDK/PDK 齐平发布的**维护成本巨大**（6×2=12 套 SDK 要同步 API） | ① **定义** **`liboriginlang`** **C ABI 为唯一真源**，SDK 都是薄绑定（除 Go 内核本身和 TS 纯 JS 实现）；② 写一套契约测试 (Contract Test)，所有 SDK 必须通过相同的 golden JSON 用例 |
| R2 | Wasm GC 语言（Python/Java）的包体积大、性能弱，用户体验不好           | ① M1/M2 文档明确标注"生产级插件推荐 Go/Rust/C++ Wasm，Py/Java 适合原型 + 子进程模式"；② M3 研究 GC-heapless Java (Chicory) 路径                                   |
| R3 | 原生动态库 (native) 模式的安全风险（段错误拉垮宿主）                   | ① 默认**禁用 native 模式**，需显式配置 `enable_native_artifacts: true` + 插件必须平台管理员白名单；② 文档中明确推荐顺序 wasm > stdio > remote > native                  |
| R4 | Web Components 与现有业务团队（React/Vue）的使用体验            | ① 提供 React/Vue 适配器 (`@originlang/react-adapter` 把 `<acme-xxx>` 包成 React 组件)；② Module Federation 模式直接允许 React/Vue 原样导出业务模块             |
| R5 | 插件市场的运营与审核成本                                      | ① M3 只做"基础市场 + 自动安全扫描"；② M4 再引入人工审核 + 商业插件支付                                                                                          |

***

## 9. 参考实现锚点（本仓库已存在的代码）

- [src/go/host/plugin/plugin.go](file:///e:/code/originlang/src/go/host/plugin/plugin.go)：插件 Spec/State/Capability 生命周期机（§3.2.1 的起点）

- [src/go/host/rpc/rpc.go](file:///e:/code/originlang/src/go/host/rpc/rpc.go)：JSON-RPC 2.0 Message 结构（§3.3 的协议基础）

- [src/go/host/transport/transport.go](file:///e:/code/originlang/src/go/host/transport/transport.go)：Transport 接口 + stdio/tcp Kind（§3.3 的传输基础）

- [src/go/host/transport/endpoint.go](file:///e:/code/originlang/src/go/host/transport/endpoint.go)：Endpoint Call/Notify + Handler 注册（§3.2 内核所有子系统共享）

- [MODULE.bazel](file:///e:/code/originlang/MODULE.bazel)：Bazel 六语言规则依赖（cc/java/rust/python/ts）（§3.1 + §5 的构建基础）

***

> **草案状态说明**：本文件为 v0.1 Draft，欢迎在 `#architecture` 频道提出建议。下一个版本将补充：
>
> 1. WIT 接口契约样例；
> 2. Security Engine 三层权限判定流程伪代码；
> 3. Manifest JSON Schema 完整字段；
> 4. M1 里程碑的任务拆分清单。

