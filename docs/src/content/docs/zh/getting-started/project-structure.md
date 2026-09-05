---
title: 项目结构
description: OriginLang 仓库结构导览。
---

OriginLang 是一个多编程语言、插件化的框架。**Bazel** 作为上层构建编排系统统一管理各语言的构建、测试与依赖。

## 顶层布局

```
originlang/
├── runtime/              # 运行时核心
│   ├── core/             # 生命周期、上下文、版本
│   ├── services/         # 系统服务：调度、权限、存储
│   ├── ipc/              # IPC：JSON-RPC、stdio、TCP
│   └── host-api/         # 插件可调用的 Host API
├── engine/               # 解释器、字节码或执行引擎
├── adapters/             # Node.js、Python、WASM 等适配层
├── sdk/                  # 各语言 SDK
├── cli/                  # 命令行工具（ol、插件脚手架）
├── developer-kit/        # 面向插件开发者的发行边界
├── modules/              # 可独立版本化的官方模块
├── plugins/              # 官方插件
├── examples/             # 示例
├── gateway/              # App / CLI / 桌面 / 远程客户端的统一入口层
├── apps/                 # 可执行应用（CLI 工具、独立程序）
├── services/             # 可部署服务 / 守护进程
├── tests/                # 测试（与运行时布局对应）
├── docs/                 # 本文档站点
├── tools/                # 构建辅助脚本、代码生成工具
└── third_party/          # vendored 外部依赖
```

## 各目录说明

| 目录 | 职责 |
| --- | --- |
| `runtime/` | 运行时核心：生命周期、上下文与版本（`core`）；调度、权限、存储等系统服务（`services`）；JSON-RPC、stdio、TCP 等 IPC（`ipc`）；以及暴露给插件的 Host API（`host-api`）。 |
| `engine/` | 执行引擎：解释器或字节码，用于求值插件逻辑。 |
| `adapters/` | 适配层，接入非核心运行时——Node.js、Python、WASM 等，使其他生态的插件也能被加载。 |
| `sdk/` | 各语言 SDK：先是参考实现的 Go SDK，随后是 Rust、Java、Python、C++、TypeScript。 |
| `cli/` | 命令行工具（`ol`）：`init` / `build` / `run` / `test` 等插件开发工作流。 |
| `developer-kit/` | 面向插件开发者的发行边界：CLI、各语言 SDK、manifest schema、模板与本地开发宿主，均独立版本化。见[开发者工具包](../reference/developer-kit/)。 |
| `modules/` | 可独立发布与版本化的官方模块。 |
| `plugins/` | OriginLang 团队维护的官方插件。 |
| `examples/` | 可运行的宿主与插件示例。 |
| `gateway/` | 统一入口层：把 HTTP、gRPC、WebSocket、本地 IPC 与 CLI 调用适配为统一的请求模型，转交宿主的稳定服务接口。见[网关](../architecture/gateway/)。 |
| `apps/` | 最终可执行入口点：CLI 工具与独立程序，每个子目录对应一个可运行程序。 |
| `services/` | 可部署的网络服务与守护进程，每个子目录可独立部署。 |
| `tests/` | 单元与集成测试，与运行时布局一一对应。 |
| `tools/` | 构建辅助脚本与代码生成工具，非 Bazel 规则管理。 |
| `third_party/` | vendored 外部依赖源码或补丁。 |

## 实现状态

上述目录定义的是实现归属边界。多个目录目前仍是骨架，因此本页说明代码应放在哪里，而不承诺某个守护进程、SDK 或 RPC 接口已经交付。只有在运行时依赖和公开契约实现后，才应在 `services/` 下新增可部署宿主。

## 依赖方向

```
apps/ · services/ · cli/ · examples/ · gateway/
                 │
                 ▼
modules/ · plugins/ · sdk/ · adapters/
                 │
                 ▼
engine/ · runtime/host-api · runtime/services · runtime/ipc
                 │
                 ▼
runtime/core
```

- 宿主依赖 `runtime/` 核心，获得插件生命周期与 IPC 能力。
- `gateway/`（与 `apps/`、`cli/` 同类）消费 `runtime/services` 与插件管理器；不得自行实现底层协议，也不加载插件。
- `sdk/`、`adapters/`、`cli/` 构建在运行时之上，供宿主消费。
- `developer-kit/` 是发行边界，而非运行时依赖：它描述 CLI、SDK、manifest schema、模板与本地开发宿主如何一起交付。
- `modules/` 与 `plugins/` 在 SDK/运行时之上打包，可独立版本化。
- `tests/` 直接验证运行时行为。

## 构建系统：Bazel 9 + Bzlmod

工作区使用现代 **Bzlmod（`MODULE.bazel`）** 管理外部依赖。由于 **Bazel 9 已移除内置语言规则**，各语言规则来自独立规则集，通过 `bazel_dep` 声明并在各 `BUILD` 文件中显式 `load()`：

- `@rules_go//go:defs.bzl` → `go_*`
- `@rules_cc//cc:defs.bzl` → `cc_*`
- `@rules_java//java:defs.bzl` → `java_*`
- `@rules_rust//rust:defs.bzl` → `rust_*`
- `@rules_python//python:defs.bzl` → `py_*`

实际用法见[使用 Bazel 构建](../guides/build-with-bazel/)。

## 配置文件

- `MODULE.bazel` — Bzlmod 外部依赖清单（各语言规则集）。
- `WORKSPACE` — 声明 `workspace(name = "originlang")`；当前主要使用 Bzlmod，预留兼容。
- `.bazelversion` — 锁定 Bazel 版本。
- `.bazelrc` — 公共构建参数。
- `.gitignore` — 忽略 Bazel 输出与 IDE/系统文件。
