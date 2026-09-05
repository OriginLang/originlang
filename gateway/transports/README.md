# Gateway Transports

此目录将适配外部入口协议，包括 HTTP、gRPC、WebSocket、本地 IPC 和 CLI。每种适配器负责解析请求、创建 Gateway 请求上下文，并将统一响应映射回该协议。

传输适配器不应实现鉴权规则、插件运行时启动或业务路由；这些分别属于 `../middleware/`、`adapters/` 与 `../routing/`。

当前仅预留目录，未实现任何监听器或协议适配器。
