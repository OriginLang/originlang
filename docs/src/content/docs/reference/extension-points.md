---
title: Extension Points
description: The contribution-point model for backend APIs and frontend UI.
---

Extension points (contribution points) are how a plugin extends the host without the host knowing the plugin in advance. The host defines **points**; plugins contribute **extensions** that attach to those points. This is what turns a set of RPC methods into a product experience — including UI.

## Model

```go
type ExtensionPoint struct {
	ID          string         // e.g. "originlang.ui.sidebar.menu"
	Description string
	Schema      JSONSchema     // the interface contract extensions must satisfy
	Scope       Scope          // Global / Tenant / User
}

type Extension struct {
	PluginID string         // which plugin contributed this
	PointID  string         // which point it attaches to
	Order    int            // sort weight
	Enabled  bool
	Config   map[string]any // extension-level configuration
}
```

## Built-in extension points

Backend points:

| Point ID | Purpose |
| --- | --- |
| `originlang.api.handler` | Plugin declares HTTP/gRPC/JSON-RPC routes; the host registers them on the gateway |
| `originlang.pipeline.hook` | Generic pipeline hooks — `before_request` / `after_response` / `event_transform` |
| `originlang.data.source` | Data-source connectors (JDBC/Redis/ES/Mongo providers) |
| `originlang.auth.provider` | Authentication providers (OIDC/LDAP/SAML/OAuth2) |
| `originlang.job.scheduler` | Scheduled jobs and DAG workflow node types |
| `originlang.host_call` | Declaration of host-provided capabilities (KV, HTTP, events, …) |

Frontend UI points:

| Point ID | Purpose |
| --- | --- |
| `originlang.ui.menu` | Inject entries into the application menu tree |
| `originlang.ui.route` | Lazy-load frontend routes via Module Federation |
| `originlang.ui.component` | Register Web Components into host views |
| `originlang.ui.setting_panel` | Plugin-specific settings pages |
| `originlang.ui.status_bar` | Status-bar widgets (desktop/IDE style) |

## How hosts consume them

1. **Startup**: the host pulls all extensions for `originlang.ui.*` points from the kernel registry.
2. **Filtering**: UI extensions declare `required_permissions`; the host filters by the current user. Tenants without an authorized plugin get no UI at all.
3. **Rendering**: menu/route/component extensions are rendered framework-agnostically — Web Components for portable UI plus Module Federation for heavy business modules.

Every component talks to the backend through a standard host object exposed by the SDK:

```ts
window.OriginLangHost.call(method, params).then((result) => { /* ... */ });
```

## Relationship to manifests

Extensions are declared in the plugin `manifest.extensions` array. The manifest binds a plugin to the points it contributes; the registry binds those contributions to the running host. See [Plugin Manifest](manifest/).

## Security

Deny-by-default applies to extension points too:

1. Only declared `extensions` are registered (nothing implicit).
2. UI extensions are filtered per-user and per-tenant.
3. Calls from UI components go through the same authorization chain as any other plugin call.