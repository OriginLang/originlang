# CLI

命令行工具。

这里实现项目初始化、构建、运行、测试和发布等开发者工作流。CLI 编排 `runtime/`、`sdk/`、`modules/` 与 `plugins/`，不重复实现其底层能力。

## 插件脚手架（规划）

开发者通过 `ol plugin init` 创建一个可构建、可测试、可由宿主加载的最小插件工程。该命令只生成代码与配置；插件运行时的加载、进程通信和权限执行仍分别属于 `adapters/`、`runtime/ipc/` 与 `runtime/services/`。

```text
ol plugin init [directory]
  → 选择插件 ID、语言、运行时载体与贡献点
  → 生成 manifest、入口、RPC bootstrap、测试与构建配置
  → 校验生成结果符合模板声明的 OriginLang 版本
```

目录规划见：

| 目录 | 职责 |
| --- | --- |
| `commands/plugin/` | `plugin init`、`build`、`test`、`pack` 等面向用户的命令编排。 |
| `scaffold/` | 参数校验、模板选择、文件计划、冲突检测和生成结果校验。 |
| `templates/plugin/` | 随 CLI 发布的、按语言和执行载体分类的插件模板。 |

当前 CLI 与脚手架仅定义目录和文档，尚未实现命令。

## Developer Kit 交付

面向插件开发者的 Developer Kit 由 CLI、各语言 SDK、manifest schema、模板、契约测试工具与本地开发宿主共同组成。CLI 负责将这些能力编排为统一工作流，但插件仍使用语言生态的原生构建器：Node 使用 npm/pnpm，Python 使用 uv/Poetry/pip，Java 使用 Maven/Gradle。

`ol plugin build` 的职责是调用模板声明的构建步骤、检查 manifest 的入口产物、运行 OriginLang 契约测试并形成插件包；它不取代语言构建系统。完整发行边界见 `../developer-kit/README.md`。

## 命令名

当前文档继续使用 `ol` 作为暂定 CLI 命令名。`origin` 已被多个现有开发者工具作为可执行命令使用，不能直接替换；若要更名，应另行选择并验证一个可分发的命令名（例如完整名称 `originlang`），同时处理包名、PATH 冲突与商标检索。
