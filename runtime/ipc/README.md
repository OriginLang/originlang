# IPC

JSON-RPC 协议及其 stdio、TCP 等传输实现。

这里负责报文、编解码、连接与传输抽象；插件生命周期和权限决策仍属于 `runtime/core` 与 `runtime/services`。
