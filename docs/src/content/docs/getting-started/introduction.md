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

## Current status

The repository currently provides the directory skeleton for the runtime (`core`, `services`, `ipc`, and `host-api`), execution engine, adapters, SDKs, CLI, modules, plugins, and examples. These boundaries guide implementation; concrete language SDKs and deployable hosts are added as their contracts land. Start with the [Project Structure](project-structure/) tour.
