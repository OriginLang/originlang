# OriginLang

OriginLang 是一个面向多语言插件系统的平台基础项目。仓库按运行时、执行引擎、语言适配与开发者工具划分职责，避免把实现代码、可发布模块和示例混在同一层。

```text
originlang/
├── runtime/              # 运行时核心
│   ├── core/             # 生命周期、上下文、版本
│   ├── services/         # 调度、权限、存储等系统服务
│   ├── ipc/              # JSON-RPC、stdio、TCP
│   └── host-api/         # 插件可调用的 Host API
├── engine/               # 解释器、字节码或执行引擎
├── adapters/             # Node.js、Python、WASM 等适配层
├── sdk/                  # 各语言 SDK
├── cli/                  # 命令行工具
├── modules/              # 可独立版本化的官方模块
├── plugins/              # 官方插件
├── examples/             # 示例
└── docs/                 # 文档站点
```

完整的目录职责与依赖方向见 [docs/structure.md](docs/structure.md)；面向使用者的文档位于 [docs/README.md](docs/README.md)。
