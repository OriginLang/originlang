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
