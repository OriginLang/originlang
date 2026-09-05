---
title: 使用 Bazel 构建
description: 用 Bazel 构建、测试、管理多语言工作区。
---

OriginLang 使用 **Bazel 9** 作为顶层构建系统。所有语言通过同一个 hermetic 工作区完成构建、测试与依赖解析。

## 工作区基础

工作区由以下文件定义：

- `MODULE.bazel` — Bzlmod 依赖清单。
- `WORKSPACE` — 声明 `workspace(name = "originlang")`（保留兼容；Bzlmod 为主）。
- `/BUILD` — 顶层包，导出工作区文件。

Bazel 9 **没有内置语言规则**。各语言规则集在 `MODULE.bazel` 中声明：

```bazel
bazel_dep(name = "rules_cc",    version = "0.2.22")
bazel_dep(name = "rules_java",  version = "9.9.0")
bazel_dep(name = "rules_rust",  version = "0.73.0")
bazel_dep(name = "rules_python", version = "1.9.2")
bazel_dep(name = "rules_go",    version = "0.63.0")
```

每个 `BUILD` 文件 `load()` 它需要的规则，例如：

```bazel
load("@rules_go//go:defs.bzl", "go_library")

go_library(
    name = "core",
    srcs = glob(["core/**/*.go"]),
    importpath = "originlang/runtime/core",
    visibility = ["//visibility:public"],
)
```

## 常用命令

```sh
# 构建整个工作区
bazel build //...

# 构建当前工作区骨架
bazel build //...

# 列出当前已声明 BUILD 文件的包
bazel query //...
```

运行时子目录在包含源码前不会声明构建 target。例如添加 `runtime/ipc/BUILD` target 后，可用 `bazel build //runtime/ipc:<target-name>` 构建它。

## 添加新的二进制或服务

1. 在 `apps/<name>/`（或 `services/<name>/`）下创建源码与 `BUILD` 文件。
2. 声明合适的规则——`cc_binary` / `java_binary` / `py_binary` / `go_binary` 等。
3. `deps` 指向你消费的运行时库，例如 `//runtime/ipc`。

## 添加新的语言模块

1. 在对应目录下添加源码，例如 `sdk/<lang>/`。
2. 在该目录的 `BUILD` 中声明库 target。
3. 在宿主中通过 `deps = ["//sdk/<lang>"]` 引用。

## 启用 TypeScript/JS

TypeScript 已搭好脚手架但尚未接入。启用时取消 `MODULE.bazel` 中 aspect 规则的注释：

```bazel
bazel_dep(name = "aspect_rules_js", version = "2.1.2")
bazel_dep(name = "aspect_rules_ts", version = "3.1.0")
```

然后在相应 `BUILD` 中使用 `ts_project`。

## 网络与代理

首次构建会从 GitHub Releases 下载上述规则集。若网络无法直连 GitHub，请配置代理：

```sh
# PowerShell
$env:HTTP_PROXY  = "http://proxy:port"
$env:HTTPS_PROXY = "http://proxy:port"
```

或写入用户级 `~/.bazelrc`：

```
common --repo_env=HTTPS_PROXY=http://proxy:port
common --repo_env=HTTP_PROXY=http://proxy:port
```
