---
title: 简介
description: OriginLang 是什么，以及为什么存在。
---

OriginLang 是一个**支持多语言、插件化的通用平台底座**。它提供了可复用的基础层，让各类可扩展产品——SaaS 平台、开发者工具、AI 工作台、企业中台、跨平台应用——都能在其上搭建。

它本身不是成品，也不引入新的编程语言。名称中的 "Lang" 指代可用于编写插件与嵌入宿主的多种主流语言。

## 支持哪些插件载体？

四种插件载体共存，共用一套生命周期：

| 载体 | 机制 | 隔离级别 | 典型延迟 |
| --- | --- | --- | --- |
| **Wasm 沙箱**（推荐） | Wasmtime 加载 WASI/Component Module 模块 | 最强 | ~1µs |
| **子进程** | 任意可执行文件，通过 stdio/TCP 说话 JSON-RPC | OS 进程级 | ~100µs |
| **原生动态库** | `dlopen` 加载 C ABI 的 `.so` / `.dll` / `.dylib` | 同进程（最弱） | ~10ns |
| **远程服务** | 部署在其他服务器或 K8s Pod 上的插件 | 网络级 | ~1ms |

## 哪些语言是一等公民？

六种语言享有同等优先级的地位——任何 SDK 的版本都必须六语言齐平发布：

**Go · Rust · Java · Python · C++ · TypeScript**

- **Go**：内核参考实现，零第三方依赖。
- **TypeScript**：所有 UI 扩展作者以及 Web/Node 宿主。
- **Rust / Java / Python / C++**：各自生态的宿主集成与插件。

## 它解决什么问题？

手工打造一个可插拔平台成本高昂：你需要设计插件协议、安全模型、多租户隔离、扩展点注册表、热更新、分发与可观测性——并且要为每种支持的语言重复一遍。OriginLang 开箱即用地提供这些：

- 与传输无关的 **JSON-RPC 2.0 协议**（stdio、TCP、未来 HTTP 与 Wasm IPC）；
- 带加载、握手、查找、调用、关闭的**插件管理器与生命周期状态机**；
- **Deny-by-Default 安全**：能力清单 → 租户授权 → 用户策略；
- 同时覆盖后端 API 与前端 UI 的**扩展点（贡献点）注册表**；
- 租户级可见性、配额、计量的**多租户**内建支持；
- 指标、日志、追踪的**零配置可观测性**；
- 一个 **Bazel 多语言构建系统**，在同一工作区构建/测试全部六种语言。

## 它刻意不做的事

- 不做任何垂直业务产品（不做 IDE、不做 CRM、不做 AI Agent 平台）。
- 不发明新的编程语言。
- 不重复造成熟基础轮子：传输用 JSON-RPC 2.0、Wasm 用 Wasmtime、编排用 K8s Operator、构建用 Bazel。

## 与同类项目的对比

现有插件平台各留短板——SDK 长尾冻结、载体单一、无多租户。OriginLang 正是对这些不足的直接回应：

| 维度 | Extism | wasmCloud | Tauri v2 | Eclipse Theia | PF4J | TEN | OriginLang |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 插件可选语言 | 10+（长尾冻结） | 5（Java 早期） | 仅 Rust | 仅 TS/JS | 仅 JVM | C++/Go/Python | 6 语言齐平发布（Go/Rust/Java/Python/C++/TS） |
| 宿主可选语言 | 16+（长尾冻结） | 主要是 Rust/Go CLI | 仅 Rust | 仅 Node | 仅 Java | 主要是 C++/Go | 6 语言齐平 |
| 前端 UI 扩展 | 无 | 无 | 有（但必须 Rust 后端） | 有（仅 TS/IDE 场景） | 无 | 无 | Web Components + MF 跨框架一体化 |
| 四种插件载体并存 | 仅 Wasm | 仅 Wasm 组件 | 编译期绑定 Rust crate | npm 包（编译期） | 仅 jar（JVM） | 原生模块 | Wasm/子进程/原生库/远程四种统一生命周期 |
| 运行时热加载 | 有 | 有 | 无（编译期绑定） | 部分（dev 模式） | 有 | 有 | 有，且四种载体都支持 |
| 细粒度 Capability 权限 | 仅 Wasm 级 | WASI deny-by-default | 仅命令级 | 仅 workspace 级 | 无内建 | 简单 | 三层：Manifest + 租户授权 + 用户 RBAC + 配额 |
| 一等公民多租户 | 无 | 无 | 无 | 无 | 无 | 无 | 内建：可见性/配额/计费/实例隔离 |
| 插件级可观测性 | 无 | 基础 tracing | 无 | 较少 | 无 | 实时场景强 | 零配置指标/日志/追踪/慢插件剖析 |
| 插件依赖版本管理 | 需自建 | 组件组合，无 SAT | 由 Cargo 解决 | npm 依赖地狱 | 简单 + Maven 冲突 | 无 | SAT 求解器 + semver + 灰度 + 回滚 |
| 部署形态 | Server/Edge/CLI/IoT | Cloud/Edge K8s | 桌面/移动 App | Browser/Electron | JVM Server | Server/SDK | 全部覆盖：单机/K8s/桌面(Tauri)/移动/边缘/IDE(Theia) |
| 定位 | 通用 Wasm 插件系统 | K8s 级 Wasm 微服务平台 | 跨平台应用框架 | IDE/开发工具平台 | Java 服务端模块化 | 实时 AI Agent 框架 | 通用多语言插件化平台底座（所有场景） |

## 当前状态

仓库当前提供运行时（`core`、`services`、`ipc`、`host-api`）、执行引擎、适配器、SDK、CLI、模块、插件、示例、[网关](../architecture/gateway/)与[开发者工具包](../reference/developer-kit/)的目录骨架。这些边界用于指导后续实现；具体语言 SDK 与可部署宿主会随其公开契约一起落地。请先阅读[项目结构](project-structure/)。
