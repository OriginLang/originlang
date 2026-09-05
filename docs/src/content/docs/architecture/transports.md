---
title: Transports
description: The pluggable transport matrix and how transports are selected.
---

A `Transport` is responsible for moving JSON-RPC messages over some medium. Transports are **symmetric**: the same interface is implemented by the host kernel and by plugins, so a plugin can run as a local child process (stdio/TCP) or as a remote service (HTTP) without changing the plugin-manager logic.

## Transport implementations

| Kind | Status | Medium | Typical latency | Isolation |
| --- | --- | --- | --- | --- |
| `stdio` | ✅ Implemented | child process stdin/stdout, newline-delimited JSON | ~100µs | OS process |
| `tcp` | ✅ Implemented | TCP socket (Dial/Listen/Server) | ~200µs | Network |
| `http` | 🚧 Planned | remote microservice, JSON-RPC over HTTPS | ~1ms | Network |
| `wasm` | 🚧 Planned | in-process Wasmtime sandbox, shared-memory IPC | ~1µs | Wasmtime sandbox |
| `native` | 🚧 Planned | in-process dynamic library via `dlopen` (C ABI) | ~10ns | Same process |
| `inproc` | 🚧 Planned | same-process direct call (host and plugin language match) | <1ns | None |

The `transport.Kind` enum already reserves `stdio` and `tcp`; the remaining kinds are added as implementations land.

## How a transport is selected

When a manifest provides multiple artifacts, the kernel selects a transport automatically by priority:

```
inproc (same language)  →  native (matching platform .so/.dll)  →  wasm (default, cross-language + safe)
              ↘ unavailable or disabled        ↘ no native ABI       ↗ stdio (universal, any executable)  →  remote (HTTP)
```

The recommended order for production is **wasm > stdio > remote > native**, with `native` disabled by default because a segfault in a loaded library can crash the host.

## The transport interface

```go
type Transport interface {
	Kind() Kind
	Send(*rpc.Message) error   // write one message
	Read() (*rpc.Message, error) // read one message
	Close() error
	String() string
}
```

Implementations tend to share a small `Codec` that handles newline-delimited JSON framing so message encoding stays identical across mediums.

## The endpoint session

Whatever the medium underneath, both host and plugin run a `transport.Endpoint` on top of their transport. The endpoint adds:

- Method handler registration (`Register`)
- Request/response correlation for concurrent calls (`Call`)
- One-way notifications (`Notify`)
- Graceful close and drain for shutdown

This is what lets a single protocol serve every carrier: from the cheapest in-process call to a remote server on another continent.