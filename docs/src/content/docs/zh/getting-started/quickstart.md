---
title: 快速开始
description: 几分钟内了解 OriginLang。
---

理解 OriginLang 最快的途径，是先看清各层职责以及未来宿主与插件共享的协议契约。

## 前置条件

- **Bazel 9**（由 `.bazelversion` 锁定）用于构建工作区骨架。
- 阅读协议示例不需要其他环境——线格式就是纯 JSON。

## 查看工作区

```sh
bazel build //...
bazel query //...
```

运行时源码与测试 target 将按[项目结构](project-structure/)中的目录边界逐步添加。

## 核心思想

一个完整的 OriginLang 系统其实就是**通过某种传输相连的两个 JSON-RPC 2.0 对等端**：

- **宿主**：作为应用或可部署服务实现，负责加载、跟踪并转发给插件；
- **插件**：独立进程，通过 stdio（或 TCP）应答 JSON-RPC。

宿主与插件无需知道彼此的语言，只要对齐 JSON-RPC 2.0 即可。

## 插件握手

宿主加载插件时发送 `plugin.register`，插件回复身份与能力：

```json
// host → plugin
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }

// plugin → host
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet" } ] } }
```

## 建议的宿主接口

宿主实现 `runtime/host-api` 契约后，客户端可使用类似接口：

```json
// 列出已加载插件
{ "jsonrpc": "2.0", "id": 1, "method": "host.plugin.list", "params": {} }

// 转发调用到某个插件
{ "jsonrpc": "2.0", "id": 2, "method": "host.plugin.call",
  "params": { "plugin_id": "plugin-1", "method": "hello.greet", "params": {} } }
```

宿主会执行查找和生命周期检查，通过插件传输转发请求，再中继结果。这些消息用于说明规划中的契约，并不是可直接对当前仓库运行的命令。

## 一个工作系统的结构

```
+-----------+   JSON-RPC 2.0 over stdio/TCP   +-----------+
|   host    | <------------------------------> |  plugin   |
| (service  |   plugin.register/ping/shutdown  |  (any     |
|  or app)  |   + custom methods / capabilities |  language)|
|           |                                  |           |
+-----------+                                  +-----------+
      ▲
      │ host.plugin.list / host.plugin.call
      ▼
  your client / UI / gateway
```

## 下一步

- 查看[项目结构](project-structure/)，熟悉仓库布局。
- 阅读[架构](../architecture/overview/)，看各部分如何协同。
- 开始[编写一个插件](../guides/write-a-plugin/)。
