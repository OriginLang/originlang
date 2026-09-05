# SDK

各语言 SDK。

每个语言子目录提供符合该生态习惯的 Host API 与插件开发接口，并依赖稳定的 `runtime/` 契约。SDK 不承载运行时核心实现或官方插件。

SDK 是面向插件开发者发行的 Developer Kit 的一部分：TypeScript SDK 应通过 npm、Python SDK 通过 PyPI、Java SDK 通过 Maven 仓库等各语言惯用渠道发布。SDK 提供类型、插件 RPC bootstrap、manifest 辅助工具和契约测试支持；最终宿主提供的 Electron、JRE、CPython 等运行时不应被打进 SDK 或插件包。

Developer Kit 的完整组成与本地开发流程见 `../developer-kit/README.md`。
