---
title: 开发者工具包
description: 面向插件开发者的发行边界。
---

开发者工具包（Developer Kit）定义插件开发者的发行边界：第三方开发者无需取得完整宿主源码、Electron 应用或内置 JRE/Python 运行时，也能创建、构建、测试与分发插件。它由多个可独立版本化的组件构成，而不是一个要求所有语言共用的构建系统。

## 组成

| 提供物 | 发行渠道 | 职责 |
| --- | --- | --- |
| CLI | 独立二进制或 Node 包 | 初始化、构建编排、测试、打包、本地开发与发布。 |
| TypeScript SDK | npm | Node/TypeScript 插件的类型、RPC bootstrap 与测试工具。 |
| Python SDK | PyPI | Python 插件的类型、RPC bootstrap 与测试工具。 |
| Java SDK | Maven 仓库 | Java/JVM 插件的类型、RPC bootstrap 与测试工具。 |
| Manifest schema | 随 CLI 与 SDK 发布 | IDE 补全、静态校验以及能力、运行时和贡献点声明。 |
| 插件模板 | 随 CLI 发布 | `plugin init` 生成的最小可构建工程。 |
| 本地开发宿主 | 由 CLI 启动 | 模拟插件生命周期、权限、路由和 Host API，用于本地调试与契约测试。 |
| 插件包规范 | 文档与 CLI 校验器 | 定义 manifest、构建产物和资源如何形成可安装包。 |

各语言 SDK 位于仓库 `sdk/`；CLI 在 `cli/` 中编排工作流。两者都不捆绑运行时：宿主提供的 Electron、JRE 或 CPython 运行时不会被打进 SDK 或插件包。

## 开发者工作流

```text
CLI 初始化模板
  → 使用语言原生构建器编译
  → CLI 执行 manifest / 契约校验
  → 本地开发宿主加载调试
  → CLI 打包并发布
```

示意命令如下；CLI 暂定命令名为 `ol`：

```bash
ol plugin init my-plugin --language typescript --runtime node
cd my-plugin
ol plugin dev
ol plugin test
ol plugin pack
ol plugin publish
```

`ol plugin build` 不应重写 npm、uv/Poetry、Maven 或 Gradle 的构建逻辑。它读取模板或项目声明的构建步骤，验证 `originlang.plugin.json` 与入口产物，然后将其封装为符合插件包规范的结果。

## 运行时责任

插件只声明运行时需求，不携带或下载宿主运行时：

```json
{
  "runtime": { "kind": "java", "version": ">=21 <22" }
}
```

最终宿主决定如何满足该声明：Electron 宿主可用内置 Node 承载 Node 插件；桌面发行物可携带受控 JRE 或 CPython；服务端与 CLI 宿主可使用管理员配置的运行时。运行时适配与解析规则归属 `adapters/`。

## 兼容性

每个 SDK、CLI、模板与插件包都应声明目标 OriginLang 版本。CLI 在初始化、构建和安装阶段校验该范围；宿主在加载前再次校验，避免"本地可构建但目标平台不可运行"。

## 命令名状态

`origin` 不是可用的默认命令名：Cursor Origin CLI 和其他项目已占用该可执行命令。当前保留 `ol` 作为暂定名称；若需要可读性更高的替代名，以完整名称 `originlang` 作为首选候选，但需另行进行名称、包注册与商标审查。

本目录目前只说明发行与开发者体验边界，尚未包含 SDK、CLI、开发宿主或发布自动化的实现。