---
title: Developer Kit
description: The distribution boundary for plugin developers.
---

The Developer Kit defines the distribution boundary for plugin developers: everything a third party needs to create, build, test, and distribute plugins without obtaining the full host source, an Electron app, or bundled JRE/Python runtimes. It is a set of independently versioned components rather than one build system shared by every language.

## Components

| Deliverable | Distribution channel | Responsibility |
| --- | --- | --- |
| CLI | Standalone binary or Node package | Initialize, orchestrate builds, test, pack, run a local development host, publish. |
| TypeScript SDK | npm | Types, RPC bootstrap, and test tooling for Node/TypeScript plugins. |
| Python SDK | PyPI | Types, RPC bootstrap, and test tooling for Python plugins. |
| Java SDK | Maven repository | Types, RPC bootstrap, and test tooling for Java/JVM plugins. |
| Manifest schema | Shipped with the CLI and SDKs | IDE completion, static validation, and capability/runtime/contribution-point declarations. |
| Plugin templates | Shipped with the CLI | Minimal buildable projects produced by `plugin init`. |
| Local development host | Launched by the CLI | Simulates plugin lifecycle, permissions, routing, and Host API for local debugging and contract tests. |
| Plugin package spec | Docs + CLI validator | Defines how the manifest, build artifacts, and resources form an installable package. |

The per-language SDKs live in the `sdk/` repository directory; the CLI drives the workflows in `cli/`. Neither bundles a runtime: the Electron, JRE, or CPython runtimes provided by a host are never packed into an SDK or plugin package.

## Developer workflow

```text
CLI initializes a template
  → compile with the language-native builder
  → CLI runs manifest / contract validation
  → local development host loads it for debugging
  → CLI packs and publishes
```

Illustrative commands (the CLI's tentative name is `ol`):

```bash
ol plugin init my-plugin --language typescript --runtime node
cd my-plugin
ol plugin dev
ol plugin test
ol plugin pack
ol plugin publish
```

`ol plugin build` never rewrites the build logic of npm, uv/Poetry, Maven, or Gradle. It reads the build steps declared by the template or project, validates `originlang.plugin.json` and the entry artifacts, then wraps the result into a plugin package.

## Runtime responsibility

Plugins only declare a runtime requirement — they do not carry or download a host runtime:

```json
{
  "runtime": { "kind": "java", "version": ">=21 <22" }
}
```

The final host decides how to satisfy the declaration: an Electron host can use its built-in Node for Node plugins, a desktop distribution can carry a controlled JRE or CPython, and server/CLI hosts can use admin-configured runtimes. Runtime adaptation and resolution rules belong to `adapters/`.

## Compatibility

Every SDK, CLI, template, and plugin package declares a target OriginLang version. The CLI validates that range on init, build, and install; the host re-validates before loading — avoiding "builds locally but won't run on the target platform."

## Command name status

`origin` is not a usable default command name: Cursor's Origin CLI and other projects already occupy the executable. `ol` is kept as the tentative name; if a more readable alternative is wanted, the full name `originlang` is the leading candidate and requires a separate review of the name, package registration, and trademarks.

The directory currently documents the distribution and developer-experience boundary only; no SDK, CLI, development host, or publishing automation is implemented yet.