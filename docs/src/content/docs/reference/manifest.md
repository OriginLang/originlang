---
title: Plugin Manifest
description: The plugin manifest declares identity, artifacts, capabilities, and extension contributions.
---

A plugin package describes itself with a **manifest** at its root — the single source of truth the host uses to load, authorize, and render a plugin.

> The manifest schema is being finalized as part of the M1 milestone. The structure below is the design contract; comments show intent and may not yet be enforced by a validator.

## Structure

```json
{
  "$schema": "https://originlang.dev/schemas/manifest-v1.json",
  "id": "com.acme.data-connector",
  "name": "Acme Data Connector",
  "version": "1.4.2",
  "originlang_min_version": "0.3.0",
  "authors": ["team@acme.com"],
  "license": "Apache-2.0",
  "description": "Connects to Acme SaaS data sources",

  "artifacts": {
    "wasm":     "dist/plugin-component.wasm",
    "binary": {
      "linux-amd64":   "dist/plugin-linux-amd64.so",
      "darwin-arm64":  "dist/plugin-mac-arm64.dylib",
      "windows-amd64": "dist/plugin-win-amd64.dll"
    },
    "remote": {
      "default_endpoint": "https://plugin.acme.com/rpc"
    }
  },

  "dependencies": [
    { "id": "originlang.stdlib", "version": ">=0.2.0" },
    { "id": "com.example.db-pool", "version": "~2.1.0", "optional": true }
  ],

  "capabilities": {
    "rpc_methods": [
      { "method": "acme.query",  "description": "Query Acme SaaS data", "timeout_ms": 3000, "memory_mb": 128 },
      { "method": "acme.export", "description": "Export data to CSV",   "timeout_ms": 30000, "memory_mb": 512 }
    ],
    "host_calls": [
      "originlang.kv.get", "originlang.kv.set",
      "originlang.http.fetch",
      "originlang.fs.read:./data/**",
      "originlang.events.publish:acme.data.*"
    ],
    "env_vars": ["ACME_API_KEY"],
    "network_egress": [
      "https://*.acme.com/*",
      "tcp://mq.acme.com:5671"
    ]
  },

  "extensions": [
    { "point": "originlang.data.source",
      "name": "acme_source", "order": 100,
      "config_schema": { "$ref": "./config-schema.json" } },
    { "point": "originlang.ui.menu",
      "props": { "label": "Acme Data", "icon": "database", "route": "/acme" } },
    { "point": "originlang.ui.route",
      "props": { "path": "/acme", "mf_module": "./dist/ui/acme-remote.js" } }
  ]
}
```

## Sections

### Identity

| Field | Description |
| --- | --- |
| `id` | Reverse-DNS style unique plugin ID (e.g. `com.acme.thing`) |
| `name` | Human-readable display name |
| `version` | Semantic version (`semver`) |
| `originlang_min_version` | Minimum runtime version this plugin requires |
| `authors` / `license` / `description` | Metadata surfaced in the marketplace and dashboards |

### Artifacts

Declares which plugin carriers are available. The runtime picks one by the [transport selection rules](../architecture/transports/):

- `wasm` — a WASI/Component Model module (default, cross-language + safe).
- `binary` — platform-specific native libraries (`.so` / `.dylib` / `.dll`).
- `remote` — a remote HTTP JSON-RPC endpoint.

### Dependencies

Version-constrained dependencies on other plugin packages. `optional` marks best-effort dependencies. Resolution starts with semver range matching (`>=`, `~`, `^`) and is planned to grow into a full solver.

### Capabilities — deny by default

The capability block declares what the plugin **needs** and **offers**:

| Section | Meaning |
| --- | --- |
| `rpc_methods` | RPC methods the plugin serves to callers (its API surface) |
| `host_calls` | Host-provided capabilities the plugin may call (KV, HTTP, filesystem paths, event subjects) |
| `env_vars` | Environment variables the plugin may read |
| `network_egress` | Outbound network destinations the plugin may reach |

Nothing is granted unless declared. This is the first of the three security layers (manifest → tenant grant → user policy).

### Extensions

Contribution points the plugin attaches to. Backend points (`originlang.data.source`, `originlang.api.handler`, …) and UI points (`originlang.ui.menu`, `originlang.ui.route`, …) are declared uniformly, so one package carries both backend APIs and frontend components. See [Extension Points](extension-points/).

## Validation

The host validates the manifest on install: required identity fields, dependency satisfiability, artifact existence, and capability declarations. Invalid manifests are rejected before the plugin is loaded.