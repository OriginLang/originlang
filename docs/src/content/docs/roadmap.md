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

**Done when**: `ol init` → `ol build` → `ol run` → call the plugin and get a correct result; the TS plugin is loadable by the same Go host.

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
- Bazel 9 + Bzlmod workspace with per-language rules wired for Go, C++, Java, Rust, and Python.

See the [Layered Architecture](architecture/overview/) and [Project Structure](getting-started/project-structure/) for the implementation boundaries behind each milestone.
