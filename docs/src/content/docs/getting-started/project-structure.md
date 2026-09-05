---
title: Project Structure
description: A tour of the OriginLang repository.
---

OriginLang is a polyglot, plugin-based framework. **Bazel** acts as the top-level build orchestration system, managing builds, tests, and dependencies for every language from a single workspace.

## Top-level layout

```
originlang/
├── runtime/              # Runtime core
│   ├── core/             # Lifecycle, context, version
│   ├── services/         # System services: scheduling, permissions, storage
│   ├── ipc/              # IPC: JSON-RPC, stdio, TCP
│   └── host-api/         # Host API callable by plugins
├── engine/               # Interpreter, bytecode or execution engine
├── adapters/             # Node.js, Python, WASM and other adapter layers
├── sdk/                  # Per-language SDKs
├── cli/                  # Command-line tooling (ol, plugin scaffolding)
├── developer-kit/        # The plugin-developer distribution boundary
├── modules/              # Officially versioned modules
├── plugins/              # Official plugins
├── examples/             # Examples
├── gateway/              # Unified entry layer for apps, CLI, desktop, remote clients
├── apps/                 # Executable applications (CLI tools, standalone programs)
├── services/             # Deployable services / daemons
├── tests/                # Tests (mirrors the runtime layout)
├── docs/                 # This documentation site
├── tools/                # Build helpers, code generators
└── third_party/          # Vendored external dependencies
```

## Directory reference

| Directory | Responsibility |
| --- | --- |
| `runtime/` | The runtime core: lifecycle, context, and version (`core`); system services such as scheduling, permissions, and storage (`services`); JSON-RPC, stdio, and TCP IPC (`ipc`); and the Host API exposed to plugins (`host-api`). |
| `engine/` | The interpreter, bytecode, or execution engine that evaluates plugin logic. |
| `adapters/` | Adapter layers that integrate non-core runtimes — Node.js, Python, WASM, and more — so plugins written in other ecosystems can be loaded. |
| `sdk/` | Per-language SDKs: the reference Go SDK first, followed by Rust, Java, Python, C++, and TypeScript. |
| `cli/` | The command-line tool (`ol`) with `init` / `build` / `run` / `test` style workflows for plugin development. |
| `developer-kit/` | The plugin-developer distribution boundary: CLI, per-language SDKs, manifest schema, templates, and a local development host, each versioned independently. See the [Developer Kit](../reference/developer-kit/) page. |
| `modules/` | Officially versioned modules that can be released and versioned independently. |
| `plugins/` | Official plugins maintained by the OriginLang team. |
| `examples/` | Runnable example hosts and plugins. |
| `gateway/` | The unified entry layer: adapts HTTP, gRPC, WebSocket, local IPC, and CLI calls into one request model for the host's stable services. See the [Gateway](../architecture/gateway/) page. |
| `apps/` | Final executable entry points: CLI tools and standalone programs. Each subdirectory is one runnable program. |
| `services/` | Deployable network services and daemons. Each subdirectory is independently deployable. |
| `tests/` | Unit and integration tests, mirroring the runtime layout. |
| `tools/` | Build helper scripts and code generators, not Bazel-rule managed. |
| `third_party/` | Vendored external dependency source or patches. |

## Implementation status

The directories above define the intended ownership boundaries. Several are currently scaffolds, so this reference describes where code belongs rather than promising a particular daemon, SDK, or RPC surface. Add a deployable host under `services/` only after its runtime dependencies and public contract are implemented.

## Dependency direction

```
apps/ · services/ · cli/ · examples/ · gateway/
                 │
                 ▼
modules/ · plugins/ · sdk/ · adapters/
                 │
                 ▼
engine/ · runtime/host-api · runtime/services · runtime/ipc
                 │
                 ▼
runtime/core
```

- Hosts depend on the `runtime/` core for plugin lifecycle and IPC.
- `gateway/` (like `apps/` and `cli/`) consumes `runtime/services` and the plugin manager; it must not re-implement lower-level protocols or load plugins itself.
- SDKs (`sdk/`), adapters (`adapters/`), and the CLI (`cli/`) build on the runtime and are consumed by hosts.
- `developer-kit/` is a distribution boundary, not a runtime dependency: it describes how CLI, SDKs, manifest schema, templates, and the local development host ship together.
- `modules/` and `plugins/` are packaged on top of the SDK/runtime and can be independently versioned.
- `tests/` exercises the runtime directly.

## Build system: Bazel 9 + Bzlmod

The workspace uses modern **Bzlmod** (`MODULE.bazel`) for external dependencies. Since **Bazel 9 removed built-in language rules**, per-language rules come from independent rule sets declared via `bazel_dep` and explicitly `load()`ed in each `BUILD` file:

- `@rules_go//go:defs.bzl` → `go_*`
- `@rules_cc//cc:defs.bzl` → `cc_*`
- `@rules_java//java:defs.bzl` → `java_*`
- `@rules_rust//rust:defs.bzl` → `rust_*`
- `@rules_python//python:defs.bzl` → `py_*`

See the [Build with Bazel](../guides/build-with-bazel/) guide for practical usage.

## Configuration files

- `MODULE.bazel` — Bzlmod external dependency manifest (per-language rule sets).
- `WORKSPACE` — declares `workspace(name = "originlang")`; primarily Bzlmod, kept for compatibility.
- `.bazelversion` — pins the Bazel version.
- `.bazelrc` — common build parameters.
- `.gitignore` — ignores Bazel output and IDE/system files.
