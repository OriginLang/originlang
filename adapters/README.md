# Adapters

Node.js、Python、WASM 等适配层。

适配层将外部运行时或执行载体接入统一的 OriginLang 运行时契约。共享协议定义应留在 `runtime/ipc/`，而非在每个适配器中重复定义。
