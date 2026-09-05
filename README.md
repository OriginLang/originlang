# OriginLang

OriginLang 是一个面向多语言插件系统的平台基础项目。

```text
originlang/
├── runtime/              # 运行时公共契约
├── engine/               # 执行引擎
├── adapters/             # 语言与运行时适配
├── sdk/                  # 各语言 SDK
├── cli/                  # 开发者命令行工具
├── modules/              # 官方模块
├── plugins/              # 官方插件
├── examples/             # 示例
├── docs/                 # 公开开发者文档站点
└── docs-private/         # 仅限团队访问的设计资料
```

面向插件与宿主开发者的公开文档位于 [docs/README.md](docs/README.md)。内部架构、路线图和实现方案集中在 `docs-private/`，发布公开仓库时应将其作为独立私有仓库或私有子模块管理。
