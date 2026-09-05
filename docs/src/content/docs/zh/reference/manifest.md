---
title: 插件清单
description: 插件清单声明身份、产物、能力与扩展贡献。
---

插件包在其根目录用一个**清单（manifest）**描述自身——它是宿主用于加载、授权与渲染插件的单一事实来源。

> 清单 Schema 正在 M1 里程碑中定稿。以下结构是设计契约；注释表示意图，可能尚未被校验器强制。

## 结构

```json
{
  "$schema": "https://originlang.dev/schemas/manifest-v1.json",
  "id": "com.acme.data-connector",
  "name": "Acme Data Connector",
  "version": "1.4.2",
  "originlang_min_version": "0.3.0",
  "authors": ["team@acme.com"],
  "license": "Apache-2.0",
  "description": "Connects to Acme SaaS data sources",

  "artifacts": {
    "wasm":     "dist/plugin-component.wasm",
    "binary": {
      "linux-amd64":   "dist/plugin-linux-amd64.so",
      "darwin-arm64":  "dist/plugin-mac-arm64.dylib",
      "windows-amd64": "dist/plugin-win-amd64.dll"
    },
    "remote": {
      "default_endpoint": "https://plugin.acme.com/rpc"
    }
  },

  "dependencies": [
    { "id": "originlang.stdlib", "version": ">=0.2.0" },
    { "id": "com.example.db-pool", "version": "~2.1.0", "optional": true }
  ],

  "capabilities": {
    "rpc_methods": [
      { "method": "acme.query",  "description": "Query Acme SaaS data", "timeout_ms": 3000, "memory_mb": 128 },
      { "method": "acme.export", "description": "Export data to CSV",   "timeout_ms": 30000, "memory_mb": 512 }
    ],
    "host_calls": [
      "originlang.kv.get", "originlang.kv.set",
      "originlang.http.fetch",
      "originlang.fs.read:./data/**",
      "originlang.events.publish:acme.data.*"
    ],
    "env_vars": ["ACME_API_KEY"],
    "network_egress": [
      "https://*.acme.com/*",
      "tcp://mq.acme.com:5671"
    ]
  },

  "extensions": [
    { "point": "originlang.data.source",
      "name": "acme_source", "order": 100,
      "config_schema": { "$ref": "./config-schema.json" } },
    { "point": "originlang.ui.menu",
      "props": { "label": "Acme Data", "icon": "database", "route": "/acme" } },
    { "point": "originlang.ui.route",
      "props": { "path": "/acme", "mf_module": "./dist/ui/acme-remote.js" } }
  ]
}
```

## 各区块

### 身份

| 字段 | 说明 |
| --- | --- |
| `id` | 反域名式唯一插件 ID（如 `com.acme.thing`） |
| `name` | 人类可读的显示名 |
| `version` | 语义化版本（`semver`） |
| `originlang_min_version` | 该插件所需的最低运行时版本 |
| `authors` / `license` / `description` | 在市场与面板中展示的元数据 |

### 产物

声明可用的插件载体。运行时根据 manifest 选择兼容载体：

- `wasm`——WASI/Component Model 模块（默认，跨语言且安全）。
- `binary`——平台特定原生库（`.so` / `.dylib` / `.dll`）。
- `remote`——远程 HTTP JSON-RPC 端点。

### 依赖

对其他插件包的版本约束依赖。`optional` 标记尽力而为的依赖。解析从 semver 范围匹配（`>=`、`~`、`^`）起步，规划升级为完整求解器。

### 能力——deny by default

能力区块声明插件**需要**什么、**提供**什么：

| 区块 | 含义 |
| --- | --- |
| `rpc_methods` | 插件对外服务的 RPC 方法（它的 API 面） |
| `host_calls` | 插件可调用的宿主能力（KV、HTTP、文件路径、事件主题） |
| `env_vars` | 插件可读取的环境变量 |
| `network_egress` | 插件可达的出网目的地 |

不声明即不授予。这是三层安全（清单 → 租户授权 → 用户策略）的第一层。

### 扩展

插件挂载的贡献点。后端点（`originlang.data.source`、`originlang.api.handler`……）与 UI 点（`originlang.ui.menu`、`originlang.ui.route`……）统一声明，因此一个包同时携带后端 API 与前端组件。见[扩展点](extension-points/)。

## 校验

宿主在安装时校验清单：必填身份字段、依赖可满足性、产物存在性与能力声明。非法清单在加载前即被拒绝。
