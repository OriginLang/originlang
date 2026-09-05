---
title: 插件生命周期
description: 插件生命周期状态机与状态迁移。
---

每个插件实例都会走过一个定义良好的生命周期。状态由 `host.plugin.list` 上报，并驱动管理器决策——例如"能否转发调用？""该关停了？"

## 状态机

```
created → starting → registered → ready → running → stopped
                              ↘ failed
```

| 状态 | 含义 |
| --- | --- |
| `created` | 已根据 `Spec` 构造实例；尚未运行任何东西 |
| `starting` | 正在拉起进程/连接 |
| `registered` | 插件已应答 `plugin.register` 握手并上报能力 |
| `ready` | 插件已就绪可供服务 |
| `running` | 服务 RPC 的稳态 |
| `stopped` | 干净关闭完成 |
| `failed` | 启动失败或崩溃，插件不可用 |

## 迁移

1. **created → starting**——开始拉起（本地插件 spawn 进程，远程插件建立连接）。
2. **starting → registered**——`plugin.register` 握手成功完成。
3. **registered → ready**——插件标记为可接受调用。
4. **ready → running**——插件开始服务流量。
5. **→ stopped**——干净处理完 `plugin.shutdown`。
6. **→ failed**——握手超时、启动错误或崩溃。failed 的插件不可调用，若是子进程则被杀死并回收。

## 握手失败

加载插件使用有界握手窗口（`RegisterTimeout`，默认 10 秒）。若插件未在期限内应答 `plugin.register`，或回复格式错误，插件转入 `failed`，加载调用返回错误。

## 关闭语义

- 宿主发送 `plugin.shutdown`，给插件排空在途工作并干净退出的机会。
- 宿主等待一段有界排空时间，随后强制杀掉（`SIGKILL`）未退出的子进程。
- 当前 `shutdown` 是通知（不期待回复），因此宿主以进程回收收尾，而非等待 RPC 响应。

## 规划中的补充

内核路线图针对热更新与流量管理增加两个状态：

- `unloading`——插件在热升级后正在退役。
- `draining`——停止新流量，等待在途调用完成后关停。

## 观测状态

`host.plugin.list` 返回每个插件的当前状态字符串，客户端（或 UI）可据此渲染插件面板：

```json
{ "jsonrpc": "2.0", "id": 1, "method": "host.plugin.list", "params": {} }
```

```json
{ "jsonrpc": "2.0", "id": 1, "result": [
  { "id": "plugin-1", "name": "hello", "state": "running",
    "capabilities": [ { "method": "hello.greet" } ] }
] }
```