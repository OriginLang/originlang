---
title: RPC 协议
description: 所有宿主–插件通信使用的 JSON-RPC 2.0 线上协议。
---

一切宿主↔插件通信都使用 **JSON-RPC 2.0**，并以**换行分隔 JSON** 分帧。协议刻意零依赖、语言中立，因此同一条消息可流经 stdio、TCP 与（未来的）HTTP。

## 消息信封

每条消息都是一个带 `jsonrpc` 字段的 JSON 对象，可含可选字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `jsonrpc` | string | 恒为 `"2.0"` |
| `id` | string / number / null | 请求与响应上存在；通知上缺席 |
| `method` | string | 请求与通知上存在 |
| `params` | JSON | 请求与通知上存在 |
| `result` | JSON | 成功响应的载荷 |
| `error` | object | 错误响应的载荷 |

消息按形态（而非类型标签）分类：

| 形态 | 种类 |
| --- | --- |
| 有 `method` 且有 `id` | 请求 |
| 有 `method`、无 `id` | 通知 |
| 无 `method`、有 `id` | 响应 |
| 带 `error` 的响应 | 错误响应 |

## 请求与响应

请求：

```json
{ "jsonrpc": "2.0", "id": 1, "method": "method.name", "params": { ... } }
```

成功响应：

```json
{ "jsonrpc": "2.0", "id": 1, "result": { ... } }
```

错误响应：

```json
{ "jsonrpc": "2.0", "id": 1,
  "error": { "code": -32601, "message": "method not found" } }
```

响应中的 `id` 必须与所应答请求的 `id` 一致，因此一条连接上可以多路复用大量并发调用。

## 通知

通知是没有 `id` 的请求。接收方处理它但不回响应：

```json
{ "jsonrpc": "2.0", "method": "plugin.shutdown" }
```

## 方法注册表

### 宿主 → 插件（协议强制）

| 方法 | 载荷 | 回复 |
| --- | --- | --- |
| `plugin.register` | `{ id, name }` | `RegisterReply`：`{ name, version, capabilities }` |
| `plugin.ping` | （无） | `{ pong: true, name }` |
| `plugin.shutdown` | （无） | 无（通知） |

`capabilities` 是 `{ method, description? }` 的对象数组，描述插件服务的 RPC 方法。

### 客户端 → 宿主（推荐的宿主接口）

| 方法 | 载荷 | 回复 |
| --- | --- | --- |
| `host.plugin.list` | 无 | `{ id, name, state, capabilities }` 数组 |
| `host.plugin.call` | `{ plugin_id, method, params }` | 插件的 result |

这是建议的 Host API 形态。可部署宿主只有在实现对应的 `runtime/host-api` 契约后才应暴露它。

## params 与 result 处理

- `params` 可缺席、为 `null`、为对象或数组。
- result 与 params 载荷会被原样透传（不重序列化），避免宿主与插件在跨语言互操作时丢失精度。
- 缺席或为 `null` 的 `params` 视为"无参数"，而非错误。

## 分帧

消息序列化为单行，以 `\n` 结尾。这让 stdio 管道上的分帧极其简单，也方便调试（`printf` / `nc`）。

## 传输无关性

所有传输上的 JSON-RPC 信封完全一致。只有分帧媒介不同：

| 传输 | 分帧媒介 |
| --- | --- |
| stdio | stdin/stdout 上的换行分隔 JSON |
| TCP | 套接字上的换行分隔 JSON |
| HTTP（未来） | HTTP POST 的 JSON body |

具体支持的传输由宿主发行物决定。
