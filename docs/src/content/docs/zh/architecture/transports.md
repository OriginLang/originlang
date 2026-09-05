---
title: 传输层
description: 可插拔传输矩阵与传输选择方式。
---

`Transport` 负责让 JSON-RPC 消息跨越某种媒介。传输是**对称的**：宿主内核与插件实现同一接口，因此插件既可作为本地子进程（stdio/TCP），也可作为远程服务（HTTP）运行，而无需修改插件管理器逻辑。

## 传输实现

| Kind | 状态 | 媒介 | 典型延迟 | 隔离 |
| --- | --- | --- | --- | --- |
| `stdio` | ✅ 已实现 | 子进程 stdin/stdout，换行分隔 JSON | ~100µs | OS 进程 |
| `tcp` | ✅ 已实现 | TCP 套接字（Dial/Listen/Server） | ~200µs | 网络 |
| `http` | 🚧 规划中 | 远程微服务，HTTPS 上的 JSON-RPC | ~1ms | 网络 |
| `wasm` | 🚧 规划中 | 进程内 Wasmtime 沙箱，共享内存 IPC | ~1µs | Wasmtime 沙箱 |
| `native` | 🚧 规划中 | 进程内动态库 `dlopen`（C ABI） | ~10ns | 同进程 |
| `inproc` | 🚧 规划中 | 同进程直接调用（宿主与插件语言一致） | <1ns | 无 |

`transport.Kind` 枚举已预留 `stdio` 与 `tcp`；其余 Kind 随实现落地逐步加入。

## 传输如何被选择

当一个清单提供多种可执行文件时，内核按优先级自动选择传输：

```
inproc (同语言)  →  native (有匹配平台 .so/.dll)  →  wasm (默认，跨语言+安全)
              ↘ 不可用或禁用        ↘ 无原生 ABI      ↗ stdio (通用，任何可执行)  →  remote (HTTP)
```

生产环境的推荐顺序是 **wasm > stdio > remote > native**，`native` 默认禁用——因为加载的库一旦段错误会拖垮宿主进程。

## 传输接口

```go
type Transport interface {
	Kind() Kind
	Send(*rpc.Message) error    // 写一条消息
	Read() (*rpc.Message, error) // 读一条消息
	Close() error
	String() string
}
```

各实现往往共享一个小 `Codec` 处理换行分隔 JSON 分帧，保证各媒介的消息编码完全一致。

## 端点会话

无论底层媒介是什么，宿主与插件都在自己的传输之上运行一个 `transport.Endpoint`。端点额外提供：

- 方法 handler 注册（`Register`）
- 并发调用的请求/响应关联（`Call`）
- 单向通知（`Notify`）
- 优雅关闭与排空

这正是让单一协议服务所有载体的原因：从最便宜的进程内调用，到另一块大陆上的远程服务器。