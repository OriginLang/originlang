# Rust framework core.
# Provides the Rust-side plugin framework (e.g. safe sandboxing, memory management).

load("@rules_rust//rust:defs.bzl", "rust_library")

package(default_visibility = ["//visibility:public"])

rust_library(
    name = "core",
    srcs = glob(["core/**/*.rs"], allow_empty = True),
    edition = "2021",
)
