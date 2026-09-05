---
title: Build with Bazel
description: Build, test, and manage the polyglot workspace with Bazel.
---

OriginLang uses **Bazel 9** as its top-level build system. All languages build, test, and resolve dependencies through one hermetic workspace.

## Workspace basics

The workspace is defined by:

- `MODULE.bazel` — the Bzlmod dependency manifest.
- `WORKSPACE` — declares `workspace(name = "originlang")` (kept for compatibility; Bzlmod is primary).
- `/BUILD` — the top-level package, exporting workspace files.

Bazel 9 has **no built-in language rules**. Per-language rule sets are declared in `MODULE.bazel`:

```bazel
bazel_dep(name = "rules_cc",    version = "0.2.22")
bazel_dep(name = "rules_java",  version = "9.9.0")
bazel_dep(name = "rules_rust",  version = "0.73.0")
bazel_dep(name = "rules_python", version = "1.9.2")
bazel_dep(name = "rules_go",    version = "0.63.0")
```

Each `BUILD` file `load()`s the rules it needs, e.g.:

```bazel
load("@rules_go//go:defs.bzl", "go_library")

go_library(
    name = "core",
    srcs = glob(["core/**/*.go"]),
    importpath = "originlang/runtime/core",
    visibility = ["//visibility:public"],
)
```

## Common commands

```sh
# Build the whole workspace
bazel build //...

# Build the current workspace skeleton
bazel build //...

# List packages that currently declare BUILD files
bazel query //...
```

The runtime subdirectories do not declare build targets until they contain source code. For example, after adding a `runtime/ipc/BUILD` target, build it with `bazel build //runtime/ipc:<target-name>`.

## Adding a new binary or service

1. Create `apps/<name>/` (or `services/<name>/`) with your sources and a `BUILD` file.
2. Declare the appropriate rule — `cc_binary` / `java_binary` / `py_binary` / `go_binary` etc.
3. Set `deps` on the runtime libraries you consume, e.g. `//runtime/ipc`.

## Adding a new language-core module

1. Add sources under the appropriate directory, e.g. `sdk/<lang>/`.
2. Declare a library target in that directory's `BUILD` file.
3. Reference it from hosts with `deps = ["//sdk/<lang>"]`.

## Enabling TypeScript/JS

TypeScript is scaffolded but not yet wired in. To enable it, uncomment the aspect rules in `MODULE.bazel`:

```bazel
bazel_dep(name = "aspect_rules_js", version = "2.1.2")
bazel_dep(name = "aspect_rules_ts", version = "3.1.0")
```

Then use `ts_project` in the relevant `BUILD` files.

## Networking / proxies

The first build downloads the rule sets from GitHub Releases. If your network cannot reach GitHub directly, configure a proxy:

```sh
# PowerShell
$env:HTTP_PROXY  = "http://proxy:port"
$env:HTTPS_PROXY = "http://proxy:port"
```

or, per user, in `~/.bazelrc`:

```
common --repo_env=HTTPS_PROXY=http://proxy:port
common --repo_env=HTTP_PROXY=http://proxy:port
```
