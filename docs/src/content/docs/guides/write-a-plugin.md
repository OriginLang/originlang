---
title: Write a Plugin
description: Build a plugin that any OriginLang host can load.
---

A plugin is an **independent process that speaks JSON-RPC 2.0 over stdio**. The host launches it, performs a registration handshake, and then forwards calls to it. The protocol is the contract — nothing ties a plugin to the host's implementation language.

## The plugin skeleton

Use the Go SDK (the reference implementation) so the protocol boilerplate is handled for you. A minimal plugin:

```go
package main

import (
	"context"

	"originlang/runtime/sdk"
	"originlang/runtime/sdk/plugin"
)

func main() {
	ctx := context.Background()
	_ = sdk.Serve(ctx, sdk.Options{
		Name:    "hello",
		Version: "0.1.0",
		Capabilities: []plugin.Capability{
			{Method: "hello.greet", Description: "Greet a user"},
		},
	})
}
```

`Serve` does the heavy lifting:

1. Establishes the stdio transport on stdin/stdout.
2. Registers the three protocol-mandated handlers: `plugin.register`, `plugin.ping`, `plugin.shutdown`.
3. Waits for the host to shut it down, or for the context to be cancelled.

The `Capabilities` list is what the plugin reports during the `plugin.register` handshake — it describes the RPC methods this plugin serves.

:::caution[Package paths]
The repository is being reorganized into the `runtime/` layout; the exact import paths above settle when the runtime packages land under `runtime/`. The wire protocol below is the stable contract.
:::

## What a compliant plugin must handle

| Method | Purpose |
| --- | --- |
| `plugin.register` | Handshake — report name, version, capabilities |
| `plugin.ping` | Health check — confirm the plugin is alive |
| `plugin.shutdown` | Graceful stop — drain and exit |

## Transport details

- **Framing**: each JSON-RPC message is written as a single line, newline-terminated.
- **Stdio**: the host writes requests to the plugin's stdin; the plugin writes responses to its stdout.
- **Stderr** is free for logging and is not part of the protocol.
- **Notifications** (no `id`) need no response.

## The conversation

The host launches your plugin, then the handshake happens:

```json
{ "jsonrpc": "2.0", "id": 1, "method": "plugin.register",
  "params": { "id": "plugin-1", "name": "hello" } }
```

```json
{ "jsonrpc": "2.0", "id": 1,
  "result": { "name": "hello", "version": "0.1.0",
    "capabilities": [ { "method": "hello.greet" } ] } }
```

After that the host may ping periodically and forward calls:

```json
{ "jsonrpc": "2.0", "id": 5, "method": "hello.greet", "params": { "name": "world" } }
```

```json
{ "jsonrpc": "2.0", "id": 5, "result": { "message": "Hello, world" } }
```

## Testing your plugin without a host

You can exercise the protocol directly by piping JSON lines into the plugin executable:

```sh
printf '{"jsonrpc":"2.0","id":1,"method":"plugin.register","params":{"id":"p","name":"hello"}}\n' \
  | ./hello-plugin
```

The plugin should print its `RegisterReply` as a JSON-RPC response line. For richer interaction, wrap the pipe with a small script that sends `plugin.ping` next.

## Lifecycle from the plugin's point of view

```
host launches plugin executable
        │
        ▼
  plugin.register  →  host stores name/version/capabilities
        │
        ▼
  plugin.ping      →  host health checks (repeated)
        │
        ▼
  plugin.shutdown  →  plugin drains and exits; host reaps the process
```

## Writing a plugin in any other language

Because the contract is JSON-RPC over stdio, you can implement a plugin in any language that can read stdin/write stdout and produce/consume JSON. No SDK is required — just handle the three protocol methods and reply to method calls.

## Next steps

- See how the host loads your plugin: [Run a Host](run-a-host/).
- Understand the wire protocol in detail: [RPC Protocol](../reference/rpc-protocol/).
- Learn the [plugin lifecycle states](../reference/plugin-lifecycle/).