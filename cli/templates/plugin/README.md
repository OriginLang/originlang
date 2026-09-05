# Plugin Templates

此目录将按语言和执行载体保存 `ol plugin init` 使用的最小插件模板。

建议的后续布局：

```text
plugin/
├── node/
│   └── stdio/       # TypeScript/Node 插件，JSON-RPC over stdio
├── python/
│   └── stdio/       # Python 插件，JSON-RPC over stdio
├── java/
│   └── stdio/       # JAR 插件，JSON-RPC over stdio
└── wasm/
    └── component/   # Wasm Component 插件
```

每个模板都应包含：`originlang.plugin.json`、最小 activate/call/shutdown 实现、契约测试、构建配置和面向开发者的 README。模板通过变量填充插件 ID、名称、版本、贡献点与目标运行时版本。

模板不得内嵌 Electron、JRE、CPython 或 Node 的二进制文件，也不得硬编码用户机器路径。运行时由最终宿主按 `adapters/README.md` 的策略提供。

当前仅为目录和模板约定，尚未加入具体模板。
