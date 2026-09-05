---
title: 命令行工具
description: OriginLang 插件开发命令行工具。
---

`ol` 是 OriginLang 的开发者工具——对标 wasmCloud 的 `wash`、Extism 的 `xtp`。它负责脚手架、构建、测试、发布与运行插件。

> CLI 随 M1 里程碑发布；下列命令为规划的界面。

## 安装

```sh
bazel build //cli
# 二进制位于 bazel-bin/cli/ol（或加入 PATH）
```

## 用法

```
ol <command> [options]

Commands:
  init      scaffold a new plugin project
  build     build a plugin package
  test      test a plugin against a managed host
  push      publish a plugin package
  install   install a plugin into a host
  run       run a local host with the given plugins
  dev       watch & reload during development
```

### `ol init`

创建一个新插件项目：

```sh
ol init --name acme-data --language go --kind process
```

生成清单、插件骨架与 `BUILD` 文件。

### `ol build`

从当前目录构建一个插件包：

```sh
ol build
```

产出清单中声明的产物（`wasm`、`binary`，或两者）。

### `ol test`

把插件加载进短期宿主并验证：

```sh
ol test
```

启动管理器、加载插件、运行 `plugin.ping`，并校验声明的能力。

### `ol run`

运行一个加载了若干插件的本地宿主，暴露 JSON-RPC 面：

```sh
ol run -plugins ./my-plugin -tcp :7600
```

### `ol dev`

监听源码变更，改动后自动重建 / 重载插件——插件作者的快速内环。

### `ol push` / `ol install`

`push` 把插件包发布到 OCI 注册中心；`install` 把插件注册到宿主与租户。两者都属于市场（M3）里程碑。

## 全局参数

| 参数 | 说明 |
| --- | --- |
| `-v` | verbose 日志 |

宿主相关命令共用：

| 参数 | 说明 |
| --- | --- |
| `-plugins` | 逗号分隔的插件可执行路径 |
| `-tcp` | TCP 监听地址（如 `:7600`）；为空表示 stdio |