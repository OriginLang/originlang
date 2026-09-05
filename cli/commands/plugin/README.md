# Plugin Commands

此目录将实现 `ol plugin` 命令组，提供创建、构建、测试、打包和本地运行插件的开发者工作流。

## `ol plugin init` 规划

`ol plugin init [directory]` 应以交互式选择和非交互参数两种方式收集以下信息：

- 插件 ID、名称、版本和描述；
- 开发语言（首批为 TypeScript/Node、Python、Java）；
- 执行载体（Node、Python、JVM、Wasm；由模板支持范围决定）；
- 贡献点，例如 API handler、数据源或 UI 菜单；
- 目标 OriginLang 版本。

该命令调用 `../../scaffold/` 选择模板并生成工程。它不得自行实现插件 RPC 协议、下载解释器/JRE，或绕过 manifest 校验。

## 生成后的最小工程

```text
my-plugin/
├── originlang.plugin.json  # 身份、运行时要求、能力与贡献点
├── src/                    # 插件业务实现与 RPC bootstrap
├── tests/                  # activate / call / shutdown 的契约测试
├── README.md               # 本地开发、构建、测试与运行说明
└── <build files>           # 语言生态所需的包或构建配置
```

模板可以按语言增加惯用文件，但以上边界保持一致。生成结果不可包含实际的 JRE、Python、Node 二进制或任何宿主私有路径。

## 后续命令

| 命令 | 责任 |
| --- | --- |
| `ol plugin build` | 调用模板声明的构建器并验证 manifest 中的入口产物存在。 |
| `ol plugin test` | 启动受控测试宿主，运行协议与能力契约测试。 |
| `ol plugin pack` | 生成带 manifest、产物与必要资源的可分发插件包。 |
| `ol plugin run` | 用本地开发宿主加载插件，供调试而非生产部署。 |

当前仅为命令规划，尚未实现。
