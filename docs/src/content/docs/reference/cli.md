---
title: CLI
description: The OriginLang command-line tool for plugin development.
---

The `ol` command is OriginLang's developer tool — the counterpart to tools like `wash` (wasmCloud) or `xtp` (Extism). It scaffolds, builds, tests, publishes, and runs plugins.

> The CLI ships as part of the M1 milestone; commands below are the planned surface.

## Installation

```sh
bazel build //cli
# binary lands in bazel-bin/cli/ol (or add it to your PATH)
```

## Usage

```
ol <command> [options]

Commands:
  init      scaffold a new plugin project
  build     build a plugin package
  test      test a plugin against a managed host
  push      publish a plugin package
  install   install a plugin into a host
  run       run a local host with the given plugins
  dev       watch & reload during development
```

### `ol init`

Scaffold a new plugin project:

```sh
ol init --name acme-data --language go --kind process
```

Generates a manifest, a plugin skeleton, and a `BUILD` file.

### `ol build`

Build a plugin package from the current directory:

```sh
ol build
```

Produces the artifacts declared in the manifest (`wasm`, `binary`, or both).

### `ol test`

Load a plugin into a short-lived host and verify it:

```sh
ol test
```

Starts a manager, loads the plugin, runs `plugin.ping`, and validates the declared capabilities.

### `ol run`

Run a local host with one or more plugins, exposing the JSON-RPC surface:

```sh
ol run -plugins ./my-plugin -tcp :7600
```

### `ol dev`

Watch source files and rebuild / reload plugins on change — the fast inner loop for plugin authors.

### `ol push` / `ol install`

`push` publishes a plugin package to an OCI registry; `install` registers a plugin with a host and tenant. Both are part of the marketplace (M3) milestone.

## Flags

Global flags:

| Flag | Description |
| --- | --- |
| `-v` | verbose logging |

Host commands share:

| Flag | Description |
| --- | --- |
| `-plugins` | comma-separated plugin executable paths |
| `-tcp` | TCP listen address (e.g. `:7600`); empty means stdio |