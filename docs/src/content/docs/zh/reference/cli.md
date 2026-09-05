---
title: 命令行工具
description: OriginLang 插件开发命令行工具。
---

`ol` 是 OriginLang 的开发者工具——对标 wasmCloud 的 `wash`、Extism 的 `xtp`。它负责脚手架、构建、测试、打包与运行插件。

> CLI 随 M1 里程碑发布；下列命令为规划的界面。该目录目前只定义边界，尚未实现任何命令。见[开发者工具包](developer-kit/)了解 CLI、SDK、模板与本地开发宿主如何一起交付。

## 安装

```sh
bazel build //cli
# 二进制位于 bazel-bin/cli/ol（或加入 PATH）
```

## 用法

插件命令组为 `ol plugin`：

```
ol plugin <command> [options]

Commands:
  init      scaffold a plugin project from a template
  dev       watch & reload during development
  build     build a plugin package from the declared build steps
  test      test a plugin against a managed host
  pack      produce a distributable plugin package
  run       load a plugin in a local development host for debugging
  publish   publish a plugin package to the marketplace (M3)
```

### `ol plugin init`

创建一个新插件项目：

```sh
ol plugin init my-plugin --language typescript --runtime node
```

收集插件 ID/名称/版本/描述、开发语言（首批为 TypeScript/Node、Python、Java）、执行载体、贡献点与目标 OriginLang 版本，然后生成最小工程。该命令只生成代码与配置——加载、IPC 与权限分别仍属于 `adapters/`、`runtime/ipc/` 与 `runtime/services/`。

生成后的最小工程：

```text
my-plugin/
├── originlang.plugin.json  # 身份、运行时要求、能力与贡献点
├── src/                    # 插件业务实现与 RPC bootstrap
├── tests/                  # activate / call / shutdown 契约测试
├── README.md
└── <build files>           # 语言生态构建配置
```

生成结果不可包含 JRE、Python、Node 二进制，也不得硬编码宿主私有路径。

### `ol plugin build`

从当前目录构建一个插件包：

```sh
ol plugin build
```

运行模板或项目声明的构建步骤，验证清单中声明的入口产物存在，运行契约测试，并产出插件包。它不会取代语言原生构建器（npm、uv/Poetry、Maven、Gradle）。

### `ol plugin test`

把插件加载进短期托管宿主并验证：

```sh
ol plugin test
```

启动管理器、加载插件、运行 `plugin.ping`，并校验声明的能力。

### `ol plugin pack`

将清单、产物与必要资源按插件包规范打包为可分发插件包。

### `ol plugin run`

用本地开发宿主加载插件用于调试——非生产部署。

### `ol plugin dev`

监听源码变更，改动后自动重建 / 重载插件——插件作者的快速内环。

### `ol plugin publish`

把插件包发布到插件市场——属于 M3 里程碑。

## 命令名

`origin` 已被其他开发者工具当作可执行命令占用，因此 `ol` 是暂定命令名。若需要可读性更高的替代名，以完整名称 `originlang` 作为首选候选，但需另行进行名称、包注册与商标审查。

## 全局参数

| 参数 | 说明 |
| --- | --- |
| `-v` | verbose 日志 |

宿主相关命令共用：

| 参数 | 说明 |
| --- | --- |
| `-plugins` | 逗号分隔的插件可执行路径 |
| `-tcp` | TCP 监听地址（如 `:7600`）；为空表示 stdio |