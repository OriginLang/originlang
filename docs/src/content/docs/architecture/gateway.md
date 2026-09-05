---
title: Gateway
description: The unified entry layer for apps, CLI, desktop, and remote clients.
---

The **gateway** is the unified entry layer through which upper-layer apps, the CLI, desktop UIs, and external clients reach an OriginLang host. It adapts HTTP, gRPC, WebSocket, local IPC, or CLI calls into a transport-independent request context, then — after authorization, tenant resolution, routing, rate limiting, audit, and observability — forwards the call to the host's stable service interfaces.

The gateway does not load plugins, start Node/JRE/Python runtimes, or implement plugin business logic.

## Directory layout

| Directory | Responsibility | Status |
| --- | --- | --- |
| `contracts/` | Transport-independent request, response, identity, tenant, and error models used by the gateway internally. | Planned |
| `middleware/` | Authentication, tenant resolution, rate limiting, audit, tracing, and request policy. | Planned |
| `routing/` | Resolution and dispatch of system routes and authorized, plugin-contributed routes. | Planned |
| `transports/` | Adaptation of entry protocols: HTTP, gRPC, WebSocket, local IPC, CLI. | Planned |

## Position and boundaries

```text
App / CLI / Electron renderer / remote client
                    │
             gateway/transports
                    │
   gateway/middleware → gateway/routing
                    │
runtime/services + Plugin Manager + runtime/host-api
                    │
             adapters/* → plugin processes
```

- `runtime/ipc/` defines the JSON-RPC, stdio, and TCP protocols between host and plugins; the gateway does not duplicate them.
- `runtime/services/` owns permissions, storage, scheduling, and other domain services; the gateway only orchestrates request policy and calls them.
- `adapters/` starts Node, Python, JVM, and Wasm plugin carriers; the gateway does not know about specific runtimes.
- Hosts in `apps/` decide whether to deploy a gateway: servers can expose HTTP/gRPC, an Electron host can expose a local-only entry from the main process, and the CLI can call a local host or a remote gateway.
- `plugins/` contribute API routes through defined extension points; the gateway mounts them only when the plugin is loaded, authorized, and route-conflict-free.

## Unified request model

Every entry eventually normalizes into a transport-independent request:

```text
RequestContext(identity, tenant, permissions, trace, deadline)
  + Route(target, method, params)
  → Response(result | typed error)
```

Unified errors, deadlines, cancellation signals, and trace context must remain passable between the gateway and the plugin call. Transport-specific status codes and connection objects must not leak into the runtime service interfaces.

## Security principles

- The gateway rejects unauthenticated, unauthorized, and tenant-less requests by default.
- Plugin route registration comes from the extension declarations of verified manifests — it cannot be created dynamically by arbitrary requests.
- The gateway forwards requests based on the runtime's permission decision; a registered route never bypasses plugin capabilities or tenant grants.
- Local Electron/CLI entries and remote HTTP entries use different trust boundaries and authentication policies.

## Current status

The directory currently defines structure and responsibility boundaries only; no network listeners, routers, middleware, or protocol adapters are implemented yet.