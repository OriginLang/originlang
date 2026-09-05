---
title: Quickstart
description: Get to know OriginLang in a few minutes.
---

The quickest way to understand OriginLang is to look at the layers and the protocol contract a future host and plugin will share.

## Prerequisites

- **Bazel 9** (pinned in `.bazelversion`) to build the workspace skeleton.
- Nothing else to read the protocol examples — the wire format is plain JSON.

## Inspect the workspace

```sh
bazel build //...
bazel query //...
```

Runtime source and test targets will be added beneath the directories described in [Project Structure](project-structure/).

## The one idea

A complete OriginLang system is just **two JSON-RPC 2.0 peers over a transport**:

- the **host**, implemented as an application or deployable service, that loads, tracks, and forwards to plugins; and
- **plugins** — independent processes that answer JSON-RPC over stdio (or TCP).

The host and plugin don't need to know each other's language. They only need to agree on JSON-RPC 2.0.

## The plugin handshake

When a host loads a plugin, it sends `plugin.register` and the plugin replies with its identity and capabilities:

```json
// host → plugin
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }

// plugin → host
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet" } ] } }
```

## A proposed host surface

Once a host implements the `runtime/host-api` contract, a client can use a surface like:

```json
// list loaded plugins
{ "jsonrpc": "2.0", "id": 1, "method": "host.plugin.list", "params": {} }

// forward a call to a plugin
{ "jsonrpc": "2.0", "id": 2, "method": "host.plugin.call",
  "params": { "plugin_id": "plugin-1", "method": "hello.greet", "params": {} } }
```

The host would perform lookup and lifecycle checks, forward the request over the plugin transport, then relay the result. These messages illustrate the planned contract; they are not a command to run against the current repository.

## Structure of a working system

```
+-----------+   JSON-RPC 2.0 over stdio/TCP   +-----------+
|   host    | <------------------------------> |  plugin   |
| (service  |   plugin.register/ping/shutdown  |  (any     |
|  or app)  |   + custom methods / capabilities |  language)|
|           |                                  |           |
+-----------+                                  +-----------+
      ▲
      │ host.plugin.list / host.plugin.call
      ▼
  your client / UI / gateway
```

## Next steps

- Explore the [project structure](project-structure/) to find your way around.
- Read the [architecture](../architecture/overview/) to see how it all fits.
- Start [writing a plugin](../guides/write-a-plugin/).
