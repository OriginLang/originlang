---
title: RPC Protocol
description: The JSON-RPC 2.0 wire protocol used for all host–plugin communication.
---

All host↔plugin communication uses **JSON-RPC 2.0**, framed as **newline-delimited JSON**. The protocol is deliberately dependency-free and language-neutral, so the same messages flow over stdio, TCP, and (future) HTTP transports.

## Message envelope

Every message is a JSON object with a `jsonrpc` field and optional extras:

| Field | Type | Description |
| --- | --- | --- |
| `jsonrpc` | string | Always `"2.0"` |
| `id` | string / number / null | Present on requests and responses; absent on notifications |
| `method` | string | Present on requests and notifications |
| `params` | JSON | Present on requests and notifications |
| `result` | JSON | Present on successful responses |
| `error` | object | Present on error responses |

Message classification is by shape, not by a type tag:

| Shape | Kind |
| --- | --- |
| has `method` and `id` | request |
| has `method`, no `id` | notification |
| no `method`, has `id` | response |
| response with `error` | error response |

## Requests and responses

Request:

```json
{ "jsonrpc": "2.0", "id": 1, "method": "method.name", "params": { ... } }
```

Successful response:

```json
{ "jsonrpc": "2.0", "id": 1, "result": { ... } }
```

Error response:

```json
{ "jsonrpc": "2.0", "id": 1,
  "error": { "code": -32601, "message": "method not found" } }
```

The `id` in a response must match the `id` of the request it answers, so many concurrent calls can be multiplexed over one connection.

## Notifications

A notification is a request with no `id`. The receiver processes it but sends no response:

```json
{ "jsonrpc": "2.0", "method": "plugin.shutdown" }
```

## Method registry

### Host → plugin (protocol-mandated)

| Method | Payload | Reply |
| --- | --- | --- |
| `plugin.register` | `{ id, name }` | `RegisterReply`: `{ name, version, capabilities }` |
| `plugin.ping` | (none) | `{ pong: true, name }` |
| `plugin.shutdown` | (none) | none (notification) |

`capabilities` is an array of `{ method, description? }` objects describing the RPC methods the plugin serves.

### Client → host (recommended host surface)

| Method | Payload | Reply |
| --- | --- | --- |
| `host.plugin.list` | none | array of `{ id, name, state, capabilities }` |
| `host.plugin.call` | `{ plugin_id, method, params }` | the plugin's result |

This is a proposed Host API shape. A deployable host may expose it only after implementing the matching `runtime/host-api` contract.

## Params and result handling

- `params` may be absent, `null`, an object, or an array.
- Result and params payloads are passed through verbatim (raw JSON) so hosts and plugins never lose fidelity on re-marshalling — important for cross-language interoperability.
- A null or absent `params` decodes as "no parameters", not an error.

## Framing

Messages are serialized to a single line and terminated with `\n`. This keeps framing trivial over stdio pipes and makes debugging with `printf` / `nc` easy.

## Transport independence

The JSON-RPC envelope is identical across all transports. Only the framing medium changes:

| Transport | Framing medium |
| --- | --- |
| stdio | newline-delimited JSON over stdin/stdout |
| TCP | newline-delimited JSON over a socket |
| HTTP (future) | JSON body over HTTP POST |

Supported transport availability is determined by the host distribution.
