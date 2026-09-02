# Python framework core.
# Provides the Python-side plugin framework (e.g. dynamic loading / script engine).

load("@rules_python//python:defs.bzl", "py_library")

package(default_visibility = ["//visibility:public"])

# Register required rules before using them.
py_library(
    name = "core",
    srcs = glob(["core/**/*.py"], allow_empty = True),
)
