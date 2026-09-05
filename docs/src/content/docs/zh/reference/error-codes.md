---
title: 错误码
description: OriginLang 协议中使用的 JSON-RPC 错误码。
---

错误以 JSON-RPC 错误对象返回：`{ "code": ..., "message": ..., "data": ... }`。

## 标准 JSON-RPC 2.0 错误码

| 错误码 | 名称 | 含义 |
| --- | --- | --- |
| `-32700` | `ParseError` | 服务端收到无效 JSON |
| `-32600` | `InvalidRequest` | 发送的不是合法的 Request 对象 |
| `-32601` | `MethodNotFound` | 方法不存在 / 不可用 |
| `-32602` | `InvalidParams` | 方法参数非法 |
| `-32603` | `InternalError` | JSON-RPC 内部错误 |

## 服务端错误保留区间

`-32099` 到 `-32000` 之间的错误码保留给实现自定义的服务端错误。

## 应用错误码

`-32000..-31900` 区间的应用级错误码为宿主/插件特定，跨传输稳定。

| 错误码 | 名称 | 含义 |
| --- | --- | --- |
| `-32001` | `PluginNotFound` | 没有加载指定 ID 的插件 |
| `-32002` | `PluginStopped` | 插件没有可用端点（已关闭 / 未连接） |
| `-32003` | `Timeout` | RPC 调用超时 |
| `-32004` | `UnknownMethod` | 目标插件不服务所请求的方法 |
| `-32005` | `OperationFailed` | 远端通用操作失败 |

## 预留 / 规划错误码

随着安全、租户与配额子系统落地，规划以下应用错误码：

| 错误码 | 名称 | 含义 |
| --- | --- | --- |
| `-32006` | `Unauthorized` | 安全引擎拒绝（权限不足） |
| `-32007` | `MethodForbidden` | 方法未在插件能力中声明 |
| `-32008` | `TenantIsolated` | 跨租户访问被拒 |
| `-32009` | `QuotaExceeded` | 资源配额超限 |
| `-32010` | `CircuitBroken` | 熔断器打开 |
| `-32011` | `DependencyMissing` | 插件依赖未满足 |

## 错误对象形态

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32002,
    "message": "endpoint closed",
    "data": { ... 可选，传输特定 ... }
  }
}
```

错误对象同时充当 Go 错误：调用失败时通过常规错误路径浮出错误码与消息，载荷可携带结构化 `data`。

## 说明

- 无论哪种传输，请求未注册的方法都返回 `-32601 MethodNotFound`。
- 调用端点已关闭的插件返回 `-32002 PluginStopped`，帮助宿主区分"插件蒸发"与"插件出错应答"。