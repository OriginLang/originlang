---
title: Layered Architecture
description: The overall layered architecture of OriginLang.
---

OriginLang is organized into eight layers, from infrastructure up to application integration.

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  L8 · Application Integration Layer                                                │
│  Domain products / enterprise platforms / AI workbenches / SaaS consoles           │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L7 · Platform Governance                                                           │
│  Plugin marketplace · OCI registry · versioning/signing/upgrade/rollback           │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L6 · UI Extension Layer                                                            │
│  Web Components + Module Federation · menus / routes / views / components / panels  │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L5 · Plugin Runtime Layer                                                          │
│  ① Wasm sandbox  ② child process  ③ native .so/.dll  ④ remote HTTP/TCP             │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L4 · Polyglot SDK Layer                                                             │
│  SDK: Go | Rust | Java | Python | C++ | TypeScript                                   │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L3 · Protocol & Transport Layer                                                    │
│  JSON-RPC 2.0 · Manifest Schema · Codec · stdio / TCP / HTTP / Wasm                 │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L2 · Kernel Layer                                                                   │
│  Plugin Manager · Extension Point Registry · Event Bus · Security · Quota · Telemetry│
├─────────────────────────────────────────────────────────────────────────────────────┤
│  L1 · Infrastructure Layer                                                           │
│  Bazel build · Wasmtime engine · K8s Operator · OCI Registry · PostgreSQL / Redis   │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Kernel implementation strategy

- **Kernel core**: belongs in `runtime/core`, `runtime/services`, `runtime/ipc`, and `runtime/host-api`. The implementation language is selected per submodule without changing this public directory boundary.
- **Performance-sensitive submodules**: such as a Wasm bridge or resource-quota enforcer belong in `engine/` or `runtime/services`; they may use Rust and an FFI boundary when justified.
- **Polyglot host adaptation**: belongs in `sdk/` and `adapters/`. SDKs consume stable runtime contracts rather than importing product applications or plugins.

## Layer highlights

### L2 · Kernel (the six subsystems)

| Subsystem | Responsibility | Status |
| --- | --- | --- |
| **Plugin Manager** | Load, handshake, lookup, shutdown; instance pool, health checks | Planned runtime service |
| **Extension Point Registry** | Contribution-point model: backend APIs + frontend UI contributions | Planned runtime service |
| **Event Bus** | In-process sync pub/sub + distributed async (NATS/Redis) | Planned runtime service |
| **Security Engine** | Three layers: capability manifest → tenant grant → user policy | Planned runtime service |
| **Resource Quota** | Token-bucket QPS, concurrency, CPU time, memory limits | Planned runtime service |
| **Observability Collector** | Prometheus metrics, OpenTelemetry traces, structured logs | Planned runtime service |

### L3 · Protocol & transports

The protocol boundary is **JSON-RPC 2.0**. `runtime/ipc` owns message framing and stdio/TCP transport implementations; HTTP, Wasm, native, and in-process transports remain future extension points. See the [Transports](transports/) page for the target matrix.

### L6 · UI extension layer

Plugin authors can attach a frontend UI extension bundle to any backend plugin. UI extensions are delivered as **Web Components (Custom Elements v1 + Shadow DOM)** — the cross-framework standard — plus **Module Federation** for lazy-loading heavy business modules. Host apps are framework-agnostic (React, Vue, Angular, Svelte all work).

## Existing repository mapping

| Layer | Location |
| --- | --- |
| L2 kernel | `runtime/core`, `runtime/services`, `runtime/host-api` |
| L3 protocol & transports | `runtime/ipc` |
| L4 polyglot SDK | `sdk/` per-language SDKs |
| L4 execution engines | `engine/`, `adapters/` (Node.js, Python, WASM) |
| L7 governance | `services/` and product applications in `apps/` |
| L1 build | root `MODULE.bazel`, `BUILD` files per directory |

See [Project Structure](../getting-started/project-structure/) for the repository-level ownership rules.
