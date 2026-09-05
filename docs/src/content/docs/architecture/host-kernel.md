---
title: The Runtime Kernel
description: How the runtime core loads plugins, drives their lifecycle, and communicates over JSON-RPC.
---

The **runtime** is OriginLang's kernel: it loads plugins, drives their lifecycle, and exposes the host RPC surface. The protocol is **JSON-RPC 2.0** with newline-delimited JSON framing, so the same messages flow over stdio, TCP, or any future transport.

The protocol is language-agnostic: plugins are independent processes (or future remote services) that speak JSON-RPC, so hosts and plugins never need to share a language.

## Where things live

| Module | Responsibility |
| --- | --- |
| `runtime/core` | Lifecycle, context, and version — the plugin state machine and identity |
| `runtime/ipc` | The transport layer: JSON-RPC 2.0 messaging over stdio and TCP |
| `runtime/services` | System services built on the core: scheduling, permissions, storage |
| `runtime/host-api` | The Host API surface exposed to plugins |
| `sdk/` | Per-language plugin SDKs (the Go SDK is the reference) |

## Adapters and execution carriers

Beyond the subprocess/stdio model, other plugin carriers (Node, Python, JVM, Wasm) are started through `adapters/`. An adapter starts, stops, and monitors a plugin process and connects its stdin/stdout or another transport to the shared JSON-RPC protocol. The shared protocol definition stays in `runtime/ipc/` and is never re-implemented per adapter.

After a carrier loads a plugin, it presents the same lifecycle and call surface:

```text
load(manifest) → activate(context) → call(method, params) → ping() → shutdown()
```

Cross-process carriers use the JSON-RPC/NDJSON protocol from `runtime/ipc/` for registration, calls, health checks, and shutdown — a plugin manager only needs to pick an adapter, not understand PGlite, Python, or Java internals.

**Runtime resolution.** A plugin manifest declares only a runtime kind and version range — never an arbitrary executable path or download URL. The host resolves the runtime in this order:

1. a runtime bundled with the host (for example Electron's built-in Node);
2. an admin-configured system runtime;
3. otherwise refuse the load with a diagnostic error.

Hosts must not force every plugin to carry its own runtime binaries.

## Loading a plugin

Loading a plugin completes a four-step flow:

1. A plugin instance is created in the `created` state, then moved to `starting`.
2. A transport is established — a subprocess stdio pair for local plugins, or a TCP connection for remote ones.
3. The runtime sends the `plugin.register` request and awaits the plugin's `RegisterReply`.
4. The plugin is marked ready (then running) and becomes callable.

The same path handles an already-connected remote plugin: as long as a JSON-RPC peer is on the other end, the runtime treats it identically to a local child process.

## The handshake

```json
// host → plugin
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }

// plugin → host
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet", "description": "Greet a user" } ] } }
```

After registration the plugin reports its `Capabilities` — the RPC methods it serves. These are stored on the plugin handle for lookup and (in the future) deny-by-default authorization.

## Built-in plugin methods

| Method | Direction | Purpose |
| --- | --- | --- |
| `plugin.register` | host → plugin | Handshake: exchange identity and capabilities |
| `plugin.ping` | host → plugin | Health check — confirm the plugin is alive |
| `plugin.shutdown` | host → plugin | Graceful stop — plugin drains and exits |

## Lifecycle state machine

```
created → starting → registered → ready → running → stopped
                              ↘ failed
```

- `created` — instance constructed.
- `starting` — transport is being brought up.
- `registered` — the handshake returned capabilities.
- `ready` — the plugin signaled it can serve.
- `running` — steady state while serving RPCs.
- `stopped` — after a clean shutdown.
- `failed` — startup failure or a crash left the plugin unusable.

## Kernel subsystems

The kernel is planned as six subsystems living in `runtime/core` and `runtime/services`. The plugin manager and extension-point registry are described on the [Extension Points](../reference/extension-points/) page; the rest are covered here.

### Event bus

Two modes coexist:

- **In-process synchronous events** — a `sync.PubSub` for lightweight plugin-to-plugin communication in single-instance deployments.
- **Distributed asynchronous events** — NATS or Redis Streams with the subject naming scheme `originlang.{tenant}.{domain}.{event}`, supporting wildcard subscription and event-style filtering (CloudEvents 1.0).

### Security engine

Three layers of permission: **capability manifest → tenant grant → user policy**. On every `plugin.call` the security engine runs four checks:

1. Is the method declared in the plugin's `Capabilities`? — otherwise `MethodForbidden`;
2. Does the current tenant's allow/deny grant cover it? — otherwise `TenantIsolated`;
3. Does the current user's RBAC/ABAC policy permit it? — otherwise `Unauthorized`;
4. Is there resource-quota headroom (QPS / concurrency / CPU / memory)? — otherwise `QuotaExceeded`.

Because every call goes through these checks, an unauthorized tenant never even sees a plugin's UI (see the [Extension Points](../reference/extension-points/) security section). The planned error codes are listed on the [Error Codes](../reference/error-codes/) page.

### Resource quota

```yaml
tenant_acme.plugins.my_data_connector:
  qps: 1000
  concurrent_calls: 50
  cpu_seconds_per_hour: 3600      # Wasm / subprocess CPU time
  memory_mb_per_instance: 512     # per-instance RSS hard limit
  daily_network_mb: 10240
  timeout_ms_per_call: 5000
```

The enforcer is a high-performance token bucket with CPU-time accounting (a Rust module behind an FFI boundary). Every call entry/exit updates it atomically; on breach it returns `QuotaExceeded` and trips the circuit breaker (3 consecutive breaches → the plugin rejects all requests for 60 seconds).

### Observability

Every `Endpoint.Call` automatically emits:

- **Metrics** (Prometheus): `originlang_plugin_call_duration_ms{plugin,method,status,tenant}`, `originlang_plugin_call_total{...}`, `originlang_plugin_cpu_seconds{...}`;
- **Traces** (OpenTelemetry): a `traceparent` header injected into the RPC params so a call can be followed across plugins and hosts;
- **Structured logs**: unified fields `plugin_id / tenant_id / method / req_id / duration_ms / error_code`, automatically correlated with `trace_id`.

Slow-plugin detection: when `P95 > threshold × 3`, the host captures a CPU profile automatically (Wasmtime's profiling API for Wasm, `SIGPROF` pprof for subprocesses).

### Hot upgrade

Upgrades proceed without dropping in-flight requests:

1. Installing a new package marks `v1.4.2` as "next"; running instances stay on `v1.4.1`.
2. **Preheat** — start N new instances into `running`, but not yet accepting traffic.
3. **Traffic split** — route requests to the new version by percentage (1% → 10% → 50% → 100%).
4. **Observe** — if error rate or P95 exceeds the threshold for 5 consecutive minutes, roll back automatically.
5. **Drain** — the old version shuts down one instance at a time once its in-flight requests finish.

## Proposed host-side RPC surface

A host may expose the following methods after their `runtime/host-api` contracts are implemented:

| Method | Purpose |
| --- | --- |
| `host.plugin.list` | List loaded plugins with id, name, state, and capabilities |
| `host.plugin.call` | Forward a call to a named plugin (`plugin_id`, `method`, `params`) |

## Request/response correlation

Each side runs a message loop with a registry of handlers and a map of pending calls keyed by request ID:

- A single reader goroutine consumes messages off the transport.
- Requests dispatch to registered handlers; unregistered methods return `MethodNotFound (-32601)`.
- Concurrent `Call`s correlate responses by ID.
- Notifications (no `id`) need no response.

This is the intended shared endpoint abstraction for host and plugin peers. It has not yet been implemented in the scaffolded runtime directories.
