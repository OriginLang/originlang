---
title: Run a Host
description: How a deployable host should be composed from the OriginLang layers.
---

A host is the process that loads plugins, drives their lifecycle, and exposes selected capabilities. The repository structure reserves `services/` for deployable hosts and `apps/` for product-specific executable entry points; neither contains a shipped host implementation yet.

## Compose a host from the runtime

Use the following ownership boundaries when implementing one:

1. Put lifecycle and execution context in `runtime/core`.
2. Put JSON-RPC messages and stdio/TCP transport code in `runtime/ipc`.
3. Put scheduling, authorization, storage, and other mediated capabilities in `runtime/services`.
4. Define plugin-callable contracts in `runtime/host-api`.
5. Add language-facing wrappers in `sdk/<language>/`, and carrier-specific integration in `adapters/` or `engine/`.
6. Create the deployable binary and its `BUILD` target under `services/<host-name>/`.

## Contract before process

Do not publish a command line, endpoint, or RPC method until its matching runtime contract exists. A host implementation should document:

- its supported plugin carriers and transports;
- the Host API methods it exposes and their authorization model;
- its lifecycle and shutdown guarantees; and
- the SDK versions with which it is compatible.

This keeps product entry points thin and lets multiple hosts reuse the same runtime.

## Next steps

- Keep integrations within the documented public contracts.
- Review [Project Structure](../getting-started/project-structure/) before adding a service.
