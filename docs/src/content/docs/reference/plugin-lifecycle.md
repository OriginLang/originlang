---
title: Plugin Lifecycle
description: The plugin lifecycle state machine and how states transition.
---

Every plugin instance moves through a well-defined lifecycle. States are reported by `host.plugin.list` and drive manager decisions such as "can I forward a call?" and "time to shut down?".

## State machine

```
created → starting → registered → ready → running → stopped
                              ↘ failed
```

| State | Meaning |
| --- | --- |
| `created` | Instance constructed from a `Spec`; nothing is running yet |
| `starting` | The process/connection is being brought up |
| `registered` | The plugin answered the `plugin.register` handshake and reported capabilities |
| `ready` | The plugin signaled it is ready to serve |
| `running` | Steady state while serving RPCs |
| `stopped` | Clean shutdown completed |
| `failed` | Startup failure or a crash left the plugin unusable |

## Transitions

1. **created → starting** — launch begins (process spawn for local plugins, connection setup for remote ones).
2. **starting → registered** — the `plugin.register` handshake completes successfully.
3. **registered → ready** — the plugin is flagged as ready to accept calls.
4. **ready → running** — the plugin is serving traffic.
5. **→ stopped** — a clean `plugin.shutdown` was processed.
6. **→ failed** — handshake timeout, startup error, or a crash. A failed plugin is not callable and (if a child process) is killed and reaped.

## Handshake failure

Loading a plugin uses a bounded handshake window (`RegisterTimeout`, default 10 seconds). If the plugin does not answer `plugin.register` in time, or the reply is malformed, the plugin transitions to `failed` and the load call returns an error.

## Shutdown semantics

- The host sends `plugin.shutdown` to give the plugin a chance to drain in-flight work and exit cleanly.
- The host waits a bounded drain period before force-killing (`SIGKILL`) a child process that has not exited.
- `shutdown` is a notification today (no reply expected), so the host follows up with process reaping rather than waiting for an RPC response.

## Planned additions

The kernel roadmap adds two states for hot reload and traffic management:

- `unloading` — plugin is being retired after live upgrade.
- `draining` — stopping new traffic, waiting for in-flight calls before shutdown.

## Observing state

`host.plugin.list` returns each plugin's current state as a string, which is how a client (or UI) can render a plugin dashboard:

```json
{ "jsonrpc": "2.0", "id": 1, "method": "host.plugin.list", "params": {} }
```

```json
{ "jsonrpc": "2.0", "id": 1, "result": [
  { "id": "plugin-1", "name": "hello", "state": "running",
    "capabilities": [ { "method": "hello.greet" } ] }
] }
```