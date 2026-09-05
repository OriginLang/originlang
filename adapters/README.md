# Adapters

Node.js、Python、JVM、WASM 等执行载体的适配层。

适配层将外部运行时接入统一的 OriginLang 运行时契约。其职责是根据插件声明的运行时类型启动、停止和监控插件进程，并将其标准输入/输出或其他传输方式接到共享的 RPC 协议。共享协议定义必须留在 `runtime/ipc/`，而非在每个适配器中重复定义。

## 边界

适配器只解决“如何运行某种载体”，不承载插件业务，也不包含具体产品的窗口或 IPC 实现：

| 位置 | 职责 |
| --- | --- |
| `runtime/services/execution/`（规划） | 运行时解析、插件进程生命周期和健康检查等通用契约。 |
| `adapters/node/`（规划） | 使用宿主提供的 Node.js 能力启动 Node 插件。Electron 宿主可使用 `utilityProcess` 实现隔离。 |
| `adapters/python/`（规划） | 解析并启动受宿主控制的 CPython 运行时。 |
| `adapters/jvm/`（规划） | 解析并启动受宿主控制的 JRE/JVM，并执行 JAR 插件。 |
| `adapters/wasm/`（规划） | 将 Wasm 插件接入相同的生命周期与调用契约。 |
| `plugins/<id>/` | 插件的 manifest、业务代码和运行时无关的资源，例如 PGlite、Python 或 Java 插件本身。 |

## 运行时与发行物

运行时二进制不应由每个插件携带：

- JRE 和可选的 CPython 属于具体宿主的发行物。桌面应用应在其打包资源中按平台和架构提供一份受控运行时，例如 `resources/runtimes/jre/<platform-arch>/`。
- `adapters/` 仅包含定位和启动这些运行时的代码，不存放大体积二进制文件。
- 服务端、CLI 或其他宿主可采用自己的运行时供应方式，只要满足相同的适配器契约。

建议的解析优先级是：宿主内置运行时 → 管理员显式配置的系统运行时 → 拒绝加载并返回可诊断错误。插件 manifest 只能声明运行时种类及版本范围，不能声明任意可执行命令、文件路径或下载地址。

## 统一插件调用模型

不论 Node、Python、Java 还是 Wasm，加载完成后都向上提供相同的生命周期和调用面：

```text
load(manifest) → activate(context) → call(method, params) → ping() → shutdown()
```

跨进程载体应通过 `runtime/ipc/` 定义的 JSON-RPC/NDJSON 协议完成注册、调用、健康检查和关闭。这样 Electron 的插件管理器只需要选择适配器，而不需要理解 PGlite、Python 或 Java 的具体实现。

## 当前状态

本目录当前仅记录架构边界；Node、Python、JVM、Wasm 载体和运行时解析服务均尚未实现。
