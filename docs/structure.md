# OriginLang 项目目录结构

OriginLang 按“运行时能力 → 执行与语言接入 → 开发者工具 → 可交付扩展”的边界组织。Bazel 负责跨语言的构建、测试与依赖编排；目录职责不依赖某一种实现语言。

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
├── developer-kit/        # 面向插件开发者的发行边界
├── modules/              # 可独立版本化的官方模块
├── plugins/              # 官方插件
├── examples/             # 示例
├── gateway/              # App/CLI/桌面/远程客户端的统一入口层
├── apps/                 # 产品或可执行应用入口
├── services/             # 可部署的外部服务
├── tests/                # 单元与集成测试
├── docs/                 # 文档站点
├── tools/                # 构建和代码生成辅助工具
└── third_party/          # 外部依赖或补丁
```

## 目录职责

| 目录 | 职责 | 不应放置 |
| --- | --- | --- |
| `runtime/` | 跨语言的生命周期、系统服务、IPC 与 Host API 契约。 | 某语言专属 SDK、插件业务代码。 |
| `engine/` | 解释、字节码或其他执行引擎实现。 | 传输协议与 CLI 工作流。 |
| `adapters/` | Node.js、Python、WASM 等外部运行时适配。 | 重复的核心协议定义。 |
| `sdk/` | 面向各语言的宿主和插件开发接口。 | 运行时核心实现。 |
| `cli/` | 初始化、构建、运行、测试与发布等命令行工作流。 | 运行时内部业务规则。 |
| `developer-kit/` | 插件开发者发行边界：CLI、SDK、模板与本地开发宿主。 | 运行时核心实现。 |
| `modules/` | 可独立发布和版本化的官方复用模块。 | 运行时底座。 |
| `plugins/` | 官方维护的插件实现。 | 通用基础设施。 |
| `examples/` | 聚焦场景的可运行最小示例。 | 正式发布模块的唯一实现。 |
| `gateway/` | App/CLI/桌面/远程客户端的统一入口：协议适配、中间件与路由。 | 插件业务、运行时启动。 |
| `apps/` / `services/` | 最终产品入口或可独立部署的服务。 | 共享运行时库。 |
| `tests/` | 对以上边界进行单元、集成和契约验证。 | 生产实现。 |

## 依赖方向

```text
apps/、services/、cli/、examples/、gateway/
             ↓
modules/、plugins/、sdk/、adapters/
             ↓
engine/、runtime/host-api、runtime/services、runtime/ipc
             ↓
runtime/core
```

- `runtime/core` 是最内层稳定模型，不依赖 SDK、模块、插件或应用。
- `runtime/services`、`runtime/ipc` 和 `runtime/host-api` 构建于核心模型之上。
- `engine/` 与 `adapters/` 接入执行载体；`sdk/` 负责面向语言的开发体验。
- `gateway/` 与 `apps/`、`cli/` 同类，消费运行时服务与插件管理器，不实现底层协议或加载插件。
- `developer-kit/` 是发行边界而非运行时依赖，定义 CLI、SDK、模板与本地开发宿主如何一起交付。
- `modules/`、`plugins/`、`cli/`、`apps/`、`services/` 和 `examples/` 只能向下依赖，不能反向依赖具体产品入口。

## 当前状态与新增代码

该结构目前是仓库的实现边界和落位约定；其中多个目录仍是骨架。新增实现时，先选择所属层，再在该层创建语言或功能子目录与对应的 `BUILD` 目标。不要恢复旧的 `src/<language>/` 聚合布局，也不要在文档中将尚未存在的守护进程、SDK 或 RPC 方法描述为已交付。

## 构建系统

`MODULE.bazel` 使用 Bzlmod 管理 Go、C++、Java、Rust 和 Python 的 Bazel 规则。每个目录在开始包含可构建源码时，再添加最小的 `BUILD` 文件和所需规则；TypeScript 规则保留为后续接入项。
