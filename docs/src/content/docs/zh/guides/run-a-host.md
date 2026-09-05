---
title: 运行宿主
description: 如何按 OriginLang 分层组合可部署宿主。
---

宿主是加载插件、驱动其生命周期并暴露选定能力的进程。仓库用 `services/` 放置可部署宿主，用 `apps/` 放置面向产品的可执行入口；当前尚未提供随仓库交付的宿主实现。

## 用运行时组合宿主

实现宿主时遵循以下归属边界：

1. 将生命周期和执行上下文放在 `runtime/core`。
2. 将 JSON-RPC 报文以及 stdio/TCP 传输放在 `runtime/ipc`。
3. 将调度、授权、存储等受控能力放在 `runtime/services`。
4. 在 `runtime/host-api` 定义插件可调用的契约。
5. 在 `sdk/<language>/` 添加面向语言的封装，在 `adapters/` 或 `engine/` 添加载体特定集成。
6. 在 `services/<host-name>/` 创建可部署二进制及其 `BUILD` target。

## 先定义契约，再提供进程

不要在对应运行时契约存在前发布命令行、端点或 RPC 方法。宿主实现应说明：

- 支持的插件载体和传输；
- 暴露的 Host API 方法及授权模型；
- 生命周期与关闭保证；以及
- 兼容的 SDK 版本。

这样可以保持产品入口轻量，并让多个宿主复用同一套运行时。

## 下一步

- 集成时遵循已公开的稳定契约。
- 新增服务前阅读[项目结构](../getting-started/project-structure/)。
