---
title: 运行时内核
description: 运行时内核如何加载插件、驱动生命周期并通过 JSON-RPC 通信。
---

**运行时**就是 OriginLang 的内核：它加载插件、驱动生命周期，并对外暴露宿主 RPC 面。协议为 **JSON-RPC 2.0**，使用换行分隔 JSON 分帧，因此同一条消息可以流经 stdio、TCP 或任何未来传输。

协议是语言无关的：插件是独立进程（或未来远程服务），只说 JSON-RPC，因此宿主与插件永远无需共享语言。

## 各部分所在位置

| 模块 | 职责 |
| --- | --- |
| `runtime/core` | 生命周期、上下文与版本——插件状态机与身份 |
| `runtime/ipc` | 传输层：stdio 与 TCP 上的 JSON-RPC 2.0 消息 |
| `runtime/services` | 构建在核心之上的系统服务：调度、权限、存储 |
| `runtime/host-api` | 暴露给插件的 Host API 面 |
| `sdk/` | 各语言插件 SDK（Go SDK 为参考实现） |

## 适配器与执行载体

除子进程/stdio 模型外，其他插件载体（Node、Python、JVM、Wasm）通过 `adapters/` 启动。适配器负责启动、停止并监控插件进程，将其标准输入/输出或其他传输方式接到共享的 JSON-RPC 协议。共享协议定义必须留在 `runtime/ipc/`，而非在每个适配器中重复定义。

任意载体加载完成后，都向上提供相同的生命周期和调用面：

```text
load(manifest) → activate(context) → call(method, params) → ping() → shutdown()
```

跨进程载体通过 `runtime/ipc/` 定义的 JSON-RPC/NDJSON 协议完成注册、调用、健康检查和关闭——插件管理器只需选择适配器，而不需要理解 PGlite、Python 或 Java 的具体实现。

**运行时解析**。插件 manifest 只声明运行时种类与版本范围，绝不能声明任意可执行命令、文件路径或下载地址。宿主按以下优先级解析运行时：

1. 宿主内置运行时（例如 Electron 内置的 Node）；
2. 管理员显式配置的系统运行时；
3. 否则拒绝加载并返回可诊断错误。

宿主不应强迫每个插件携带自己的运行时二进制。

## 加载插件

加载一个插件走完四步：

1. 创建插件实例，状态为 `created`，随后置为 `starting`。
2. 建立传输通道——本地插件使用子进程 stdio 对，远程插件使用 TCP 连接。
3. 运行时发送 `plugin.register` 请求，等待插件的 `RegisterReply`。
4. 插件标记为 ready（随后进入 running），变为可调用。

已连接好的远程插件也走同一路径：只要对端是一个 JSON-RPC 对等端，运行时对待它与本地子进程完全一致。

## 握手

```json
// host → plugin
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }

// plugin → host
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet", "description": "Greet a user" } ] } }
```

注册完成后，插件上报其 `Capabilities`——它所服务的 RPC 方法。这些能力保存在插件句柄上供查找和（未来）deny-by-default 授权使用。

## 内置插件方法

| 方法 | 方向 | 用途 |
| --- | --- | --- |
| `plugin.register` | host → plugin | 握手：交换身份与能力 |
| `plugin.ping` | host → plugin | 健康检查——确认插件存活 |
| `plugin.shutdown` | host → plugin | 优雅停止——插件排空并退出 |

## 生命周期状态机

```
created → starting → registered → ready → running → stopped
                              ↘ failed
```

- `created`——实例已构造。
- `starting`——正在拉起传输通道。
- `registered`——握手已返回能力。
- `ready`——插件已就绪可供服务。
- `running`——服务 RPC 的稳态。
- `stopped`——干净关闭之后。
- `failed`——启动失败或崩溃导致插件不可用。

## 内核子系统

内核规划为六大子系统，归属于 `runtime/core` 与 `runtime/services`。插件管理器与扩展点注册表见[扩展点](../reference/extension-points/)页面，其余在此说明。

### 事件总线

两种模式共存：

- **进程内同步事件**——`sync.PubSub`，用于单实例部署下插件间轻量通信；
- **分布式异步事件**——NATS 或 Redis Streams，subject 命名规范 `originlang.{tenant}.{domain}.{event}`，支持通配符订阅与事件模式过滤（CloudEvents 1.0）。

### 安全引擎

三层权限：**能力清单 → 租户授权 → 用户策略**。内核对每次 `plugin.call` 执行四项检查：

1. 方法是否在插件声明的 `Capabilities` 中？——否则 `MethodForbidden`；
2. 是否命中当前租户授权（允许/拒绝）？——否则 `TenantIsolated`；
3. 是否命中当前用户的 RBAC/ABAC 策略？——否则 `Unauthorized`；
4. 是否还有资源配额余量（QPS/并发/CPU/内存）？——否则 `QuotaExceeded`。

由于所有调用都走这些检查，未授权租户连插件的 UI 都看不到（见[扩展点](../reference/extension-points/)的安全小节）。规划中的错误码见[错误码](../reference/error-codes/)页面。

### 资源配额

```yaml
tenant_acme.plugins.my_data_connector:
  qps: 1000
  concurrent_calls: 50
  cpu_seconds_per_hour: 3600      # Wasm / 子进程 CPU 时间
  memory_mb_per_instance: 512     # 单实例 RSS 硬限
  daily_network_mb: 10240
  timeout_ms_per_call: 5000
```

执行器为高性能令牌桶 + CPU 时间计时（Rust 模块，经 FFI 边界暴露）。每次调用进入/退出时原子更新；超限返回 `QuotaExceeded` 并触发熔断（连续 3 次超限 → 插件 60 秒内拒绝所有请求）。

### 可观测性

每一次 `Endpoint.Call` 自动生成：

- **指标**（Prometheus）：`originlang_plugin_call_duration_ms{plugin,method,status,tenant}` / `originlang_plugin_call_total{...}` / `originlang_plugin_cpu_seconds{...}`；
- **追踪**（OpenTelemetry）：自动注入 `traceparent` 到 RPC params，可跨插件、跨宿主全链路追踪；
- **结构化日志**：统一字段 `plugin_id / tenant_id / method / req_id / duration_ms / error_code`，自动关联 `trace_id`。

慢插件检测：当 `P95 > 阈值 × 3` 时，宿主自动抓取 CPU Profile（Wasm 用 Wasmtime Profiling API，子进程用 `SIGPROF` pprof）。

### 热升级

升级过程不中断在途请求：

1. 安装新版本插件包，内核标记 `v1.4.2` 为"待切换"，当前运行实例仍是 `v1.4.1`。
2. **预热**——启动 N 个新实例进入 `running`，但暂不接收流量。
3. **流量切分**——按百分比（1% → 10% → 50% → 100%）把请求路由到新版本。
4. **观测**——新版本错误率 / P95 若连续 5 分钟超阈值，自动回滚。
5. **排空**——等所有在途请求结束后，老版本实例逐一 Shutdown 回收。

## 建议的宿主侧 RPC 面

宿主可在对应 `runtime/host-api` 契约实现后暴露以下方法：

| 方法 | 用途 |
| --- | --- |
| `host.plugin.list` | 列出已加载插件：id、name、state、capabilities |
| `host.plugin.call` | 转发调用到指定插件（`plugin_id`、`method`、`params`） |

## 请求/响应关联

每一侧都跑一个消息循环：一张 handler 注册表 + 一张以请求 ID 为键的 pending 调用表。

- 单个 reader goroutine 从传输通道消费消息。
- 请求分发给注册的 handler；未注册方法返回 `MethodNotFound (-32601)`。
- 并发的 `Call` 按 ID 关联响应。
- 通知（无 `id`）无需响应。

这是宿主与插件对等端共享端点抽象的目标行为；它尚未在当前运行时骨架目录中实现。
