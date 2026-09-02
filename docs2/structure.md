# originlang 项目目录结构

originlang 是一个多编程语言、插件化的框架。Bazel 作为上层构建编排系统统一管理各语言的构建、测试与依赖。

## 依赖方向图

```
        ┌─────────┐
        │  apps/  │  可执行入口（CLI 工具、独立程序）
        └────┬────┘
             │
        ┌────▼────┐
        │services/│  可部署服务（HTTP/gRPC server、守护进程）
        └────┬────┘
             │
             ▼
      ┌───────────┐
      │   src/    │  框架核心源码（按语言分目录）
      │ 各语言core │   -> src/{java,cpp,rust,python,ts}:core
      └─────┬─────┘
            │
            ▼
   ┌──────────────────┐
   │  third_party/    │  第三方依赖 / vendored 源码
   └──────────────────┘
```

- `apps/` 与 `services/` 通过 `deps = ["//src/<lang>:core"]` 依赖框架核心。
- `tests/` 与 `src/` 一一对应，针对各语言核心编写单元/集成测试。
- `third_party/` 为外部依赖，不反向依赖本仓库代码。

## 各目录说明

| 目录 | 职责 | 典型 Bazel target |
|---|---|---|
| `apps/` | 最终可执行入口点：CLI 工具、命令行程序。每个子目录对应一个可运行程序 | `cc_binary` / `java_binary` / `py_binary` |
| `services/` | 可部署的网络服务/守护进程。每个子目录对应一个可独立部署的服务 | `cc_binary` / `java_binary` 等 + 部署配置 |
| `src/` | 框架核心源码，按语言分目录。每个语言提供 `core` 库 target，是框架本体的主要实现 | `cc_library` / `java_library` / `rust_library` / `py_library` / `ts_project` |
| `tests/` | 各语言单元测试与集成测试，与 `src/` 结构对应 | `cc_test` / `java_test` / `rust_test` / `py_test` |
| `third_party/` | 外部 vendored 依赖源码或补丁 | 无（依赖通过 MODULE 声明） |
| `tools/` | 构建辅助脚本、代码生成工具等，非 Bazel 规则管理 | 无 |

## 各语言框架核心（src/<lang>）

| 语言 | 目录 | 定位 | 规则 |
|---|---|---|---|
| Java | `src/java/core` | Java 侧框架核心（插件加载器、SPI 等） | `java_library`（rules_java） |
| C++ | `src/cpp/core` | C++ 侧框架核心（高性能运行时、FFI 接口） | `cc_library`（rules_cc） |
| Rust | `src/rust/core` | Rust 侧框架核心（安全沙箱、内存管理） | `rust_library`（rules_rust） |
| Python | `src/python/core` | Python 侧框架核心（脚本引擎、动态加载） | `py_library`（rules_python） |
| TypeScript | `src/ts/core` | TS 侧框架核心（Web UI、插件市场前端） | `filegroup`（占位，待接入 aspect_rules_ts） |

> 初始阶段各 `core` 目录仅有占位文件，实际源码在后续开发中逐步填充。

## 构建系统：Bazel 9 + Bzlmod

本仓库使用现代 **Bzlmod（`MODULE.bazel`）** 方式管理外部依赖。由于 **Bazel 9 已移除内置语言规则**，C++/Java/Rust/Python 规则分别来自独立规则集，在 `MODULE.bazel` 中通过 `bazel_dep` 引入，并在各 `BUILD` 文件中显式 `load()`：

- `@rules_cc//cc:defs.bzl` → `cc_*`
- `@rules_java//java:defs.bzl` → `java_*`
- `@rules_rust//rust:defs.bzl` → `rust_*`
- `@rules_python//python:defs.bzl` → `py_*`

TypeScript 尚未接入（`src/ts` 暂用 `filegroup` 占位）。后续添加 TS 时启用：

```bazel
bazel_dep(name = "aspect_rules_js", version = "2.1.2")
bazel_dep(name = "aspect_rules_ts", version = "3.1.0")
```

再将 `src/ts/BUILD` 的 `filegroup` 替换为：

```bazel
load("@aspect_rules_ts//ts:defs.bzl", "ts_project")
ts_project(name = "core", srcs = glob(["core/**/*.ts"]), tsconfig = "//:tsconfig")
```

### 网络与代理

首次构建时，Bazel 会从 GitHub Releases 下载上述规则集（`rules_cc`、`rules_java`、`rules_rust`、`rules_python`）。若所处网络无法直连 GitHub，请预先配置代理，例如在终端设置：

```sh
# Windows PowerShell
$env:HTTP_PROXY  = "http://proxy:port"
$env:HTTPS_PROXY = "http://proxy:port"
```

或在用户级 `~/.bazelrc` 写入：

```
common --repo_env=HTTPS_PROXY=http://proxy:port
common --repo_env=HTTP_PROXY=http://proxy:port
```

## Go Host 内核（src/go）

Go 侧实现了 **host 守护进程内核**：负责插件的加载、握手、运行与生命周期管理。插件被建模为可信的独立进程（或未来的远程服务），与 host 通过 **JSON-RPC 2.0** 通信（换行分隔 JSON）。传输层可插拔，单一协议同时支持 stdio 与 TCP，并为将来经 HTTP 做分布式部署预留。

```
src/go/                       模块 originlang（零第三方依赖，纯标准库）
├─ go.mod                     模块定义（module originlang）
├─ BUILD.bazel                gazelle + :host go_library 聚合
├─ go_deps.bzl                外部 Go 依赖说明（当前为空，占位）
└─ host/
   ├─ rpc/                    JSON-RPC 2.0 信封：Message（RPC 报文）、错误码、ID 分配
   ├─ transport/              可插拔传输抽象：Transport/Kind/Codec + Endpoint（注册/调用/通知/派发）
   │  ├─ stdio/               stdio 传输（本地子进程插件默认通道）
   │  └─ tcp/                 TCP 传输（Dial/Listen/Server，远程插件、预留分布式）
   ├─ plugin/                 插件抽象：State、Capability、Spec、生命周期、握手常量
   ├─ manager/                插件管理器：加载、握手、查找、关闭、错误映射
   └─ sdk/                    插件侧 SDK：Serve() 注册 register/ping/shutdown 处理器
```

服务入口 `services/hostd/` 提供 `hostd` 守护进程：
`-tcp` 暴露 TCP 端口，`-plugins` 以逗号分隔加载插件，并暴露 `host.plugin.list` / `host.plugin.call` 方法。

### 本地运行与测试（非 Bazel）

Bazel 之外，Go 代码用 `go test` 亦可直接验证（需本机已安装 Go 1.22+）。为避免与 Bazel 依赖冲突，原生测试/构建使用各自的 `go.mod`（通过 `replace` 指向 `../../src/go`）：

```sh
# 编译内核 + 守护进程
cd src/go && go build ./...

# 运行单元测试（rpc / transport / stdio / tcp）
cd tests/go && go test ./...

# 构建 hostd（原生验证）
cd services/hostd && go build ./...
```

> 以上 `go.mod` 仅为本地开发/原生验证便利；正式构建以 Bazel（BUILD.bazel / go_test）为准。
> `src/go/BUILD.bazel` 通过 gazelle 生成，新增 Go 源文件后运行 `bazel run //:gazelle` 刷新。

## 常用命令

```sh
# 构建整个工作区
bazel build //...

# 构建某个框架核心
bazel build //src/java:core

# 运行所有测试
bazel test //...

# 查询某 target 的依赖
bazel query 'deps(//services:api_server)'
```

## 添加内容指引

**添加一个新应用**
1. 在 `apps/<name>/` 下创建 `BUILD` 和入口源文件。
2. 在 `BUILD` 中声明 `cc_binary` / `java_binary` / `py_binary` 等，`deps` 指向 `//src/<lang>:core`。

**添加一个新服务**
1. 在 `services/<name>/` 下创建 `BUILD` 和入口源文件。
2. 在 `BUILD` 中声明相应可执行 target，依赖 `//src/<lang>:core`。

**扩展某语言的核心**
1. 在 `src/<lang>/core/` 下添加源码。
2. 更新 `src/<lang>/BUILD` 中的 `glob` 或显式 `srcs` 列表即可。

## 配置文件

- `MODULE.bazel`：Bzlmod 外部依赖清单（各语言规则集）。
- `WORKSPACE`：声明 `workspace(name = "originlang")`；当前主要使用 Bzlmod，预留兼容。
- `.bazelversion`：锁定 Bazel 版本（本仓库基于 9.2.0）。
- `.bazelrc`：公共构建参数。
- `.gitignore`：忽略 Bazel 输出与 IDE/系统文件。
