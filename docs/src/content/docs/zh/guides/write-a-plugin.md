---
title: 编写插件
description: 构建一个可被任意 OriginLang 宿主加载的插件。
---

插件就是一个**通过 stdio 说 JSON-RPC 2.0 的独立进程**。宿主启动它、完成注册握手，然后再转发调用给它。协议是契约——插件与宿主的实现语言完全无关。

## 插件骨架

用 Go SDK（参考实现）帮你处理协议样板。最小插件：

```go
package main

import (
	"context"

	"originlang/runtime/sdk"
	"originlang/runtime/sdk/plugin"
)

func main() {
	ctx := context.Background()
	_ = sdk.Serve(ctx, sdk.Options{
		Name:    "hello",
		Version: "0.1.0",
		Capabilities: []plugin.Capability{
			{Method: "hello.greet", Description: "Greet a user"},
		},
	})
}
```

`Serve` 替你完成了重活：

1. 在 stdin/stdout 上建立 stdio 传输。
2. 注册三个协议强制 handler：`plugin.register`、`plugin.ping`、`plugin.shutdown`。
3. 等待宿主关闭它，或等待 context 被取消。

`Capabilities` 列表是插件在 `plugin.register` 握手中上报的内容——描述该插件服务的 RPC 方法。

:::caution[包路径]
仓库正在重组为 `runtime/` 布局；上方的 import 路径以 `runtime/` 下落地后的包为准。下面描述的线上协议才是稳定契约。
:::

## 合规插件必须处理的方法

| 方法 | 用途 |
| --- | --- |
| `plugin.register` | 握手——上报 name、version、capabilities |
| `plugin.ping` | 健康检查——确认插件存活 |
| `plugin.shutdown` | 优雅停止——排空并退出 |

## 传输细节

- **分帧**：每条 JSON-RPC 消息占一行，以换行结尾。
- **Stdio**：宿主把请求写入插件 stdin；插件把响应写到自己的 stdout。
- **Stderr** 自由用于日志，不属于协议。
- **通知**（无 `id`）无需响应。

## 一次会话

宿主启动你的插件后，握手开始：

```json
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }
```

```json
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet" } ] } }
```

此后宿主会周期性 ping，并转发调用：

```json
{ "jsonrpc": "2.0", "id": 5, "method": "hello.greet", "params": { "name": "world" } }
```

```json
{ "jsonrpc": "2.0", "id": 5, "result": { "message": "Hello, world" } }
```

## 不借助宿主测试插件

你可以在没有宿主的情况下验证插件：把 JSON 行通过管道喂给插件可执行文件。

```sh
printf '{"jsonrpc":"2.0","id":1,"method":"plugin.register","params":{"id":"p","name":"hello"}}\n' \
  | ./hello-plugin
```

插件应把 `RegisterReply` 作为一行 JSON-RPC 响应打印出来。更丰富的交互可以用小脚本接着发送 `plugin.ping`。

## 从插件视角看的生命周期

```
宿主启动插件可执行文件
        │
        ▼
  plugin.register  →  宿主存储 name/version/capabilities
        │
        ▼
  plugin.ping      →  宿主健康检查（重复）
        │
        ▼
  plugin.shutdown  →  插件排空并退出；宿主回收进程
```

## 用任何其他语言写插件

因为契约就是 stdio 上的 JSON-RPC，你可以用任何能读写 stdin/stdout 并产出/消费 JSON 的语言实现插件。不需要 SDK——只要处理三个协议方法，并应答方法调用即可。

## 下一步

- 看宿主如何加载你的插件：[运行宿主](run-a-host/)。
- 深入了解线上协议：[RPC 协议](../reference/rpc-protocol/)。
- 学习[插件生命周期状态](../reference/plugin-lifecycle/)。