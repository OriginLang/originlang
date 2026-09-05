---
title: CLI
description: The OriginLang command-line tool for plugin development.
---

The `ol` command is OriginLang's developer tool — the counterpart to tools like `wash` (wasmCloud) or `xtp` (Extism). It scaffolds, builds, tests, packages, and runs plugins.

> The CLI ships as part of the M1 milestone; commands below are the planned surface. The directory currently defines boundaries only — no commands are implemented yet. See the [Developer Kit](developer-kit/) for how the CLI, SDKs, templates, and the local development host ship together.

## Installation

```sh
bazel build //cli
# binary lands in bazel-bin/cli/ol (or add it to your PATH)
```

## Usage

The plugin command group is `ol plugin`:

```
ol plugin <command> [options]

Commands:
  init      scaffold a plugin project from a template
  dev       watch & reload during development
  build     build a plugin package from the declared build steps
  test      test a plugin against a managed host
  pack      produce a distributable plugin package
  run       load a plugin in a local development host for debugging
  publish   publish a plugin package to the marketplace (M3)
```

### `ol plugin init`

Create a new plugin project:

```sh
ol plugin init my-plugin --language typescript --runtime node
```

Collects the plugin id/name/version/description, development language (TypeScript/Node, Python, Java first), execution carrier, contribution points, and the target OriginLang version, then generates a minimal project. It only generates code and configuration — loading, IPC, and permissions stay in `adapters/`, `runtime/ipc/`, and `runtime/services/` respectively.

Generated layout:

```text
my-plugin/
├── originlang.plugin.json  # identity, runtime requirements, capabilities, contribution points
├── src/                    # plugin logic + RPC bootstrap
├── tests/                  # activate / call / shutdown contract tests
├── README.md
└── <build files>           # language-ecosystem build config
```

The generated project must not embed JRE, Python, or Node binaries, nor hard-code host-private paths.

### `ol plugin build`

Build a plugin package from the current directory:

```sh
ol plugin build
```

Runs the build steps declared by the template or project, validates that the manifest's entry artifacts exist, runs contract tests, and produces the plugin package. It does not replace the language's native builder (npm, uv/Poetry, Maven, Gradle).

### `ol plugin test`

Load a plugin into a short-lived managed host and verify it:

```sh
ol plugin test
```

Starts a manager, loads the plugin, runs `plugin.ping`, and validates the declared capabilities.

### `ol plugin pack`

Package the manifest, artifacts, and any required resources into a distributable plugin package per the plugin package spec.

### `ol plugin run`

Load a plugin with the local development host for debugging — not for production deployment.

### `ol plugin dev`

Watch source files and rebuild / reload plugins on change — the fast inner loop for plugin authors.

### `ol plugin publish`

Publish a plugin package to the marketplace — part of the M3 milestone.

## Command name

`origin` is already taken as an executable by other developer tools, so `ol` is the tentative command name. If a more readable alternative is desired, the full name `originlang` is the leading candidate and requires a separate name, package-registration, and trademark review.

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