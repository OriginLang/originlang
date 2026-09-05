---
title: 路线图
description: 从内核 MVP 到生产级生态的四个里程碑。
---

OriginLang 分四个里程碑推进。

## M1 · 内核闭环（MVP）— 0~3 个月

**目标**：端到端跑通——宿主加载插件、扩展注册、权限拦截、示例应用单机运行。

- 运行时核心：传输（今日 stdio/TCP，新增 HTTP/inproc）、插件载体、扩展点注册表。
- 安全引擎 v1（三层权限校验）与资源配额 v1。
- `ol` CLI（`init` / `build` / `test` / `run`）。
- Go + TypeScript 的 SDK。
- 一个示例应用：若干插件（Go 与 TS）、一个菜单 UI 扩展、一个 API 扩展。

**完成标志**：`ol plugin init` → `ol plugin build` → `ol plugin run` → 调用插件并得到正确结果；TS 插件可被同一个 Go 宿主加载。

## M2 · 六语言齐平 + 可观测性 — 3~6 个月

- Rust、Java、Python、C++ 的 SDK 全部达 v1.0，六语言齐平发布。
- 可观测性：Prometheus 指标、OpenTelemetry 追踪、结构化日志、慢插件剖析。
- 契约测试套件：六语言 × 传输通过同一组 golden JSON-RPC 用例。
- 前端 UI SDK v1（menu / route / component / setting_panel 四个扩展点）。
- 热升级与回滚。

**完成标志**：CI 中 6 语言 × 2 传输全绿；P95 延迟与错误率在 Grafana 大盘可见；10 个插件同时热升级零丢请求。

## M3 · 多租户 + 插件市场 — 6~9 个月

- OCI 注册中心分发 + cosign 签名。
- 插件市场 Web 应用 + 多租户管理控制台。
- 依赖冲突求解器、灰度升级、回滚。
- 分布式事件总线（NATS）与远程 HTTP 插件 + Kubernetes Operator v1。

**完成标志**：三个租户同时使用，租户间 100% 隔离；市场上 20+ 官方示例插件；push → install → 下架 → 回滚全流程可用。

## M4 · 生产级 + 生态孵化 — 9~12 个月

- 完整韧性：熔断、限流、降级、故障注入。
- 全部 10 个 UI 扩展点；Wasm WIT 绑定生成器。
- Tauri 与 Theia 集成模板。
- 性能白皮书（四大载体的延迟与吞吐基准）。

**完成标志**：至少一个真实业务团队在生产使用；3+ 个第三方 SDK 贡献；可观的社区采用。

## 现在交付了什么

仓库当前提供以下基础：

- 运行时骨架：`runtime/`（core、services、ipc、host-api），以及 `engine/`、`adapters/`、`sdk/`、`cli/`。
- [网关](architecture/gateway/)入口层与[开发者工具包](reference/developer-kit/)发行边界（`gateway/`、`developer-kit/`）。
- Bazel 9 + Bzlmod 工作区，已接入 Go、C++、Java、Rust、Python 各语言规则。

各里程碑背后的实现边界见[分层架构](architecture/overview/)与[项目结构](getting-started/project-structure/)。

## 已知挑战与关键决策

路线图由一组已确认的风险及决定的缓解决策塑造：

| # | 挑战 | 缓解决策 |
| --- | --- | --- |
| C1 | 六语言 SDK 齐平发布的维护成本——6 套 SDK 需保持 API 一致 | 定义内核 C ABI（`liboriginlang`）为唯一真源，SDK 均为薄绑定；共享契约测试套件在所有 SDK 中跑同一组 golden JSON-RPC 用例 |
| C2 | Go 嵌入 Wasmtime 的 FFI 复杂度 | MVP 先用 `wasmtime-go` 绑定包，不自写 FFI 层 |
| C3 | GC 语言（Python/Java）的 Wasm 包体积与性能 | 文档明确生产级 Wasm 插件为 Go/Rust/C++，Python/Java 适合原型 + 子进程模式；M3 研究 GC-heapless Java（Chicory）路径 |
| C4 | 原生动态库（`.so`/`.dll`）的崩溃安全——段错误拉垮宿主 | 原生 artifact 默认禁用，需显式 `enable_native_artifacts: true` + 平台管理员白名单；文档推荐顺序 wasm > stdio > remote > native |
| C5 | Web Components 对 React/Vue 团队的使用体验 | 提供 `@originlang/react-adapter`（把 Web Component 包成 React 组件）；Module Federation 允许业务模块以 React/Vue 原样交付 |
| C6 | UI 扩展层跨框架兼容性 | 以 Web Components 为标准载体，Module Federation 作为懒加载补充 |
| C7 | SAT 依赖求解器复杂度 | M1 先用拓扑排序 + semver 范围匹配，M3 升级为完整 SAT 求解器 |
| C8 | 插件市场的运营与审核成本 | M3 只做基础市场 + 自动安全扫描；人工审核与插件商业支付在 M4 引入 |
