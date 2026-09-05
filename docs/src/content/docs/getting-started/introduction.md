---
title: Introduction
description: What OriginLang is and why it exists.
---

OriginLang is a **polyglot, plugin-based platform foundation** (底座). It provides the reusable layer upon which extensible products — SaaS platforms, developer tools, AI workbenches, enterprise middle-platforms, and cross-platform apps — can be built.

It is not a product itself, and it does not introduce its own programming language. The "Lang" in OriginLang refers to the many mainstream languages you can use to write plugins and embed the host.

## Which plugin carriers are supported?

Four plugin carriers coexist under one unified lifecycle:

| Carrier | Mechanism | Isolation | Latency |
| --- | --- | --- | --- |
| **Wasm sandbox** (recommended) | Wasmtime loads a WASI/Component Model module | Strongest | ~1µs |
| **Child process** | Any executable speaking JSON-RPC over stdio/TCP | OS process level | ~100µs |
| **Native dynamic library** | `dlopen` a C-ABI `.so` / `.dll` / `.dylib` | Same process (weakest) | ~10ns |
| **Remote service** | A plugin deployed on another server or K8s pod | Network level | ~1ms |

## Which languages are first-class?

Six languages get equal priority — any SDK release ships in all six simultaneously:

**Go · Rust · Java · Python · C++ · TypeScript**

- **Go** — reference kernel implementation, dependency-free.
- **TypeScript** — all UI extension authors and web/Node hosts.
- **Rust / Java / Python / C++** — host integrations and plugins for their respective ecosystems.

## What problem does it solve?

Building a pluggable platform by hand is expensive: you must design a plugin protocol, a security model, multi-tenant isolation, an extension registry, hot reload, distribution, and observability — and then repeat all of it for every language you support. OriginLang provides these working pieces out of the box:

- A **JSON-RPC 2.0 protocol** that is transport-agnostic (stdlib, TCP, HTTP, and future Wasm IPC).
- A **plugin manager and lifecycle state machine** with load, handshake, lookup, call, and shutdown.
- **Deny-by-default security**: capability manifests → tenant grants → user policies.
- **Extension point (contribution point) registry** for both backend APIs and frontend UI contributions.
- **Multi-tenancy** with tenant-scoped visibility, quotas, and metering.
- **Zero-config observability** for metrics, logs, and traces.
- A **Bazel-based polyglot build system** that builds and tests all six languages from one workspace.

## What it deliberately does not do

- No vertical business product (no IDE, no CRM, no AI agent platform).
- No new programming language.
- No re-inventing mature infrastructure: JSON-RPC 2.0 for transport, Wasmtime for Wasm, Kubernetes Operators for orchestration, and Bazel for build.

## How it compares

Existing plugin platforms each leave holes — a frozen long-tail of SDKs, a single carrier, no multi-tenancy. Here is where OriginLang answers them:

| Dimension | Extism | wasmCloud | Tauri v2 | Eclipse Theia | PF4J | TEN | OriginLang |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Plugin languages | 10+ (long tail frozen) | 5 (Java early) | Rust only | TS/JS only | JVM only | C++/Go/Python | 6 languages in lockstep (Go/Rust/Java/Python/C++/TS) |
| Host languages | 16+ (long tail frozen) | Rust/Go CLI mostly | Rust only | Node only | Java only | C++/Go mostly | 6 languages in lockstep |
| Frontend UI extensions | No | No | Yes (Rust backend required) | Yes (TS, IDE-scoped) | No | No | Web Components + Module Federation, framework-agnostic |
| Four carriers coexist | Wasm only | Wasm components only | Compile-time Rust crate | npm package (compile-time) | jar only (JVM) | Native modules | Wasm / subprocess / native lib / remote under one lifecycle |
| Runtime hot reload | Yes | Yes | No (compile-time binding) | Partial (dev mode) | Yes | Yes | Yes, for all four carriers |
| Fine-grained capability permissions | Wasm-level only | WASI deny-by-default | Command-level only | Workspace-level only | None built in | Simple | Three layers: manifest + tenant grant + user RBAC + quota |
| First-class multi-tenancy | No | No | No | No | No | No | Built-in: visibility / quota / metering / instance isolation |
| Plugin-level observability | No | Basic traces | No | Sparse | No | Strong in real-time scenarios | Zero-config metrics / logs / traces / slow-plugin profiling |
| Dependency version management | DIY | Component composition (no SAT) | Via Cargo | npm dependency hell | Simple + Maven conflicts | None | SAT solver + semver + canary + rollback |
| Deployment shapes | Server / Edge / CLI / IoT | Cloud / Edge K8s | Desktop / mobile app | Browser / Electron | JVM server | Server / SDK | All: standalone / K8s / desktop (Tauri) / mobile / edge / IDE (Theia) |
| Positioning | General Wasm plugin system | K8s-grade Wasm microservice platform | Cross-platform app framework | IDE / dev-tool platform | Java server modularization | Real-time AI agent framework | General polyglot plugin platform foundation (all scenarios) |

## Current status

The repository currently provides the directory skeleton for the runtime (`core`, `services`, `ipc`, and `host-api`), execution engine, adapters, SDKs, CLI, modules, plugins, examples, Gateway, and the [Developer Kit](../reference/developer-kit/). Concrete language SDKs and deployable hosts are added as their public contracts land. Start with the [Project Structure](project-structure/) tour.
