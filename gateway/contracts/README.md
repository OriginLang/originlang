# Gateway Contracts

此目录将定义 Gateway 的传输无关契约：`RequestContext`、身份与租户信息、路由目标、响应信封、规范错误、截止时间和取消信号。

契约不得依赖 HTTP 框架、Electron 或具体 CLI 库；传输适配位于 `../transports/`，领域服务与插件 RPC 协议分别位于 `runtime/services/` 和 `runtime/ipc/`。

当前仅预留目录，未实现类型或协议。
