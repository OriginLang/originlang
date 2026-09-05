# Engine

解释器、字节码或执行引擎。

执行引擎关注插件代码的求值与执行，不负责语言 SDK、进程通信或宿主业务服务；这些分别属于 `sdk/`、`runtime/ipc/` 与 `runtime/services/`。
