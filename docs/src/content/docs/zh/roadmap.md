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

**完成标志**：`ol init` → `ol build` → `ol run` → 调用插件并得到正确结果；TS 插件可被同一个 Go 宿主加载。

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
- Bazel 9 + Bzlmod 工作区，已接入 Go、C++、Java、Rust、Python 各语言规则。

各里程碑背后的实现边界见[分层架构](architecture/overview/)与[项目结构](getting-started/project-structure/)。
