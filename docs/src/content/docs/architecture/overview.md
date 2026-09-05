---
title: Layered Architecture
description: The overall layered architecture of OriginLang.
---

OriginLang is organized into eight layers, from infrastructure up to application integration.

## Design principles

The eight-layer shape is driven by seven principles, each a direct response to a gap found in existing plugin platforms (Extism, wasmCloud, Tauri, Theia, PF4J, TEN):

| # | Principle | Addresses |
| --- | --- | --- |
| P1 | Language-agnostic first; all six SDKs ship in lockstep | Extism's frozen long-tail SDKs / Tauri requires Rust / Theia is TS-only / PF4J is JVM-only |
| P2 | Protocol decoupled from transport: one JSON-RPC 2.0 protocol over stdio, TCP, HTTP, and Wasm IPC | Stalled waPC / most projects ship a single deployment shape |
| P3 | Four plugin carriers coexist under one unified lifecycle | Extism/Wasm skips native performance / PF4J is JVM-only / Tauri binds at compile time |
| P4 | Frontend and backend extensions declared in one manifest | Nearly every competitor splits these into two systems |
| P5 | Deny-by-default security + fine-grained capabilities + tenant-level re-authorization | Tauri plugins share the process without isolation / Extism has no tenant concept |
| P6 | Multi-tenancy as a first-class citizen: plugin visibility, instance isolation, quota metering | No competitor has built-in multi-tenancy |
| P7 | Unified observability: plugin-level metrics, logs, and traces with zero config | Wasm-style projects are hard to debug / Theia slow extensions are hard to trace |

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

Reference-implementation conventions (from the current Go kernel):

- The Go core stays **zero third-party dependencies** (stdlib only).
- `json.RawMessage` is passed through as-is — no lossy re-marshalling — which is critical for cross-language interoperability.
- A **single symmetric `Transport` interface** is implemented by both host and plugin ends, so remote and local plugins are handled identically.
- The manager exposes dual entry points: `Load` (subprocess via stdio) and `Attach` (remote over TCP).
- The plugin SDK's `Serve()` bootstraps a compliant plugin (register/ping/shutdown handlers) in one call.

## Layer highlights

### L1 · Infrastructure

| Module | Choice | Responsibility |
| --- | --- | --- |
| Build system | Bazel 9 (`MODULE.bazel`) | Unified six-language builds, cross-platform artifacts, caching, hermetic tests |
| Wasm engine | Wasmtime 43+ + Component Model + WIT | Wasm sandbox plugin runtime (Extism/wasmCloud-class) |
| Orchestration | Kubernetes + OriginLang Operator | Plugin scheduling, scaling, rolling upgrades in distributed deployments |
| Distribution | OCI Registry (ORAS) | Plugin packages as OCI artifacts; Sigstore/cosign signing |
| Metadata storage | PostgreSQL | Plugin metadata, versions, tenant grants, quotas, audit logs |
| Cache / events | Redis + optional NATS | Event bus backends, plugin instance-pool cache, distributed locks |

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

### L7 · Platform governance

| Sub-module | Function |
| --- | --- |
| **OCI Registry** | Plugin packages as OCI artifacts (`oras push/pull`); cosign signing; semver tags + channels (`stable` / `beta` / `nightly`) |
| **Plugin marketplace web** | Browse/search/rate/review/usage stats; payments for commercial plugins; security scan reports (Wasm security audit + SBOM) |
| **Versioning & upgrade** | Automatic dependency-conflict resolution (SAT solver); canary upgrades per tenant (1% → 10% → 50% → 100%); one-click rollback |
| **Multi-tenant console** | Tenant admins: grant/disable plugins, configure quotas, view billing and usage reports; platform admins: global takedown / circuit-break |
| **Operations analytics** | Install counts, active tenants, DAU, error rate, P95 dashboards; automatic alerting to plugin authors |

**Plugin install flow.** After an admin authorizes a plugin, the governance layer pulls the package from the OCI registry, verifies signature + SBOM, parses the manifest and validates dependency compatibility, writes version + tenant-grant rows, assigns a resource-quota profile, then notifies the kernel — which lazily loads the plugin on first call.

### L8 · Application integration

OriginLang itself ships no vertical business features — this layer is left to adopters:

- **Enterprise middle-platforms** (Java/Spring Boot) with line-of-business plugins in Go/Java/Python plus UI extensions;
- **AI workbench products** (Python + TypeScript), where model providers and toolchains are OriginLang plugins;
- **Developer tools / IDEs** (Go/Rust + TypeScript) where language servers, debuggers, and linters are plugins;
- **Cross-platform desktop apps** (Rust) embedding Tauri — OriginLang owns the plugin system, Tauri owns windows and system APIs.

## Existing repository mapping

| Layer | Location |
| --- | --- |
| L2 kernel | `runtime/core`, `runtime/services`, `runtime/host-api` |
| L3 protocol & transports | `runtime/ipc` |
| L4 polyglot SDK | `sdk/` per-language SDKs |
| L4 execution engines | `engine/`, `adapters/` (Node.js, Python, WASM) |
| Client-facing entry | `gateway/` (contracts, middleware, routing, transports) |
| L7 governance | `services/` and product applications in `apps/` |
| L1 build | root `MODULE.bazel`, `BUILD` files per directory |
| Developer Kit | `developer-kit/` distribution boundary; CLI workflows in `cli/` |

See [Project Structure](../getting-started/project-structure/) for the repository-level ownership rules.
