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
