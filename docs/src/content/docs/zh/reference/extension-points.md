---
title: 扩展点
description: 后端 API 与前端 UI 的贡献点模型。
---

扩展点（contribution point）让插件无需宿主预先知晓即可扩展宿主。宿主定义**点**；插件贡献挂到这些点上的**扩展**。这是把一组 RPC 方法变成产品体验——包括 UI——的关键。

## 模型

```go
type ExtensionPoint struct {
	ID          string         // 例如 "originlang.ui.sidebar.menu"
	Description string
	Schema      JSONSchema     // 扩展必须满足的接口契约
	Scope       Scope          // Global / Tenant / User
}

type Extension struct {
	PluginID string         // 哪个插件贡献的
	PointID  string         // 挂到哪个点
	Order    int            // 排序权重
	Enabled  bool
	Config   map[string]any // 扩展级配置
}
```

## 内置扩展点

后端扩展点：

| 点 ID | 用途 |
| --- | --- |
| `originlang.api.handler` | 插件声明 HTTP/gRPC/JSON-RPC 路由；宿主自动注册到网关 |
| `originlang.pipeline.hook` | 通用流水线钩子——`before_request` / `after_response` / `event_transform` |
| `originlang.data.source` | 数据源连接器（JDBC/Redis/ES/Mongo 提供者） |
| `originlang.auth.provider` | 认证提供者（OIDC/LDAP/SAML/OAuth2） |
| `originlang.job.scheduler` | 定时任务与 DAG 工作流节点类型 |
| `originlang.host_call` | 宿主能力声明（KV、HTTP、事件……） |

前端 UI 扩展点：

| 点 ID | 用途 |
| --- | --- |
| `originlang.ui.menu` | 向应用菜单树注入条目 |
| `originlang.ui.route` | 通过 Module Federation 懒加载前端路由 |
| `originlang.ui.component` | 把 Web Component 注册进宿主视图 |
| `originlang.ui.setting_panel` | 插件专属设置页 |
| `originlang.ui.status_bar` | 状态栏小部件（桌面/IDE 风格） |

## 宿主如何消费

1. **启动时**：宿主从内核注册表拉取 `originlang.ui.*` 点的全部扩展。
2. **过滤**：UI 扩展声明 `required_permissions`；宿主按当前用户过滤。未授权该插件的租户完全拿不到任何 UI。
3. **渲染**：菜单/路由/组件扩展以框架无关方式渲染——Web Components 承载可移植 UI，Module Federation 承接重型业务模块。

每个组件都通过 SDK 暴露的标准宿主对象与后端通信：

```ts
window.OriginLangHost.call(method, params).then((result) => { /* ... */ });
```

## 与清单的关系

扩展在插件 `manifest.extensions` 数组中声明。清单把插件绑定到它贡献的点；注册表把这些贡献绑定到运行中的宿主。见[插件清单](manifest/)。

## 安全

deny-by-default 同样适用于扩展点：

1. 只有声明的 `extensions` 才被注册（没有隐式内容）。
2. UI 扩展按用户、按租户过滤。
3. 来自 UI 组件的调用走与任何插件调用相同的授权链路。