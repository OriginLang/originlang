---
title: Roadmap
description: The four milestones from kernel MVP to production-grade ecosystem.
---

OriginLang ships in four milestones.

## M1 · Kernel loop (MVP) — 0–3 months

**Goal**: a working end-to-end system — host loads plugins, extensions register, permissions enforce, and a sample app runs on a single machine.

- Runtime core: transports (stdio/TCP today, HTTP/inproc added), plugin carriers, extension-point registry.
- Security Engine v1 (three-layer permission checks) and Resource Quota v1.
- `ol` CLI (`init` / `build` / `test` / `run`).
- Go + TypeScript SDKs.
- A sample app with several plugins (Go and TS), one menu UI extension, and one API extension.

**Done when**: `ol plugin init` → `ol plugin build` → `ol plugin run` → call the plugin and get a correct result; the TS plugin is loadable by the same Go host.

## M2 · Six languages + observability — 3–6 months

- Rust, Java, Python, and C++ SDKs reach v1.0, all six shipping in lockstep.
- Observability: Prometheus metrics, OpenTelemetry traces, structured logs, slow-plugin profiling.
- Contract test suite: all six languages x transports pass the same golden JSON-RPC cases.
- Frontend UI SDK v1 (menu / route / component / settings-panel extension points).
- Hot upgrade and rollback.

**Done when**: CI shows 6 languages × 2 transports green; P95 latency and error rate visible on a Grafana dashboard; 10 plugins upgrade with zero dropped requests.

## M3 · Multi-tenancy + marketplace — 6–9 months

- OCI registry distribution with cosign signing.
- Plugin marketplace web app + multi-tenant admin console.
- Dependency conflict solver, canary upgrades, rollback.
- Distributed event bus (NATS) and remote HTTP plugins + Kubernetes Operator v1.

**Done when**: three tenants running with 100% isolation; 20+ official sample plugins; push → install → unpublish → rollback all work end to end.

## M4 · Production-grade + ecosystem — 9–12 months

- Full resilience: circuit breakers, rate limiting, degradation, fault injection.
- All 10 UI extension points; Wasm WIT binding generator.
- Tauri and Theia integration templates.
- Performance white paper (latency & throughput across the four carriers).

**Done when**: at least one real business team in production; 3+ third-party SDK contributions; meaningful community adoption.

## What is shipping now

The repository currently provides the foundation:

- The runtime skeleton: `runtime/` (core, services, ipc, host-api), `engine/`, `adapters/`, `sdk/`, `cli/`.
- The [Gateway](architecture/gateway/) entry layer and the [Developer Kit](reference/developer-kit/) distribution boundary (`gateway/`, `developer-kit/`).
- Bazel 9 + Bzlmod workspace with per-language rules wired for Go, C++, Java, Rust, and Python.

See the [Layered Architecture](architecture/overview/) and [Project Structure](getting-started/project-structure/) for the implementation boundaries behind each milestone.

## Known challenges and decisions

The roadmap is shaped by a set of acknowledged risks and the decisions made to mitigate them:

| # | Challenge | Mitigation |
| --- | --- | --- |
| C1 | Six-SDK lockstep maintenance — six SDKs must stay API-aligned | Define the kernel C ABI (`liboriginlang`) as the single source of truth; SDKs are thin bindings; a shared contract-test suite runs identical golden JSON-RPC cases in every SDK |
| C2 | Embedding Wasmtime from Go (FFI complexity) | Start with the `wasmtime-go` binding package for the MVP instead of hand-writing the FFI layer |
| C3 | GC-language Wasm size and performance (Python/Java) | Document production-grade Wasm plugins as Go/Rust/C++, with Python/Java suited to prototypes + subprocess mode; investigate GC-heapless Java (Chicory) in M3 |
| C4 | Native `.so` / `.dll` crash safety — a segfault kills the host | Native artifacts disabled by default; opt-in via `enable_native_artifacts: true` + admin whitelist; documented carrier preference is wasm > stdio > remote > native |
| C5 | Web Components adoption for React/Vue teams | Ship `@originlang/react-adapter` (wraps Web Components as React components); Module Federation lets business modules ship in React/Vue untouched |
| C6 | Cross-framework UI compatibility | Web Components as the standard carrier, Module Federation as the lazy-loading supplement |
| C7 | SAT dependency-solver complexity | M1 uses topological sort + semver range matching; upgrade to a full SAT solver in M3 |
| C8 | Marketplace moderation and audit cost | M3 ships a basic marketplace with automated security scanning; manual review and plugin commerce arrive in M4 |
