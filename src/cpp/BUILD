# C++ framework core.
# Provides the C++-side plugin framework / high-performance runtime.

load("@rules_cc//cc:defs.bzl", "cc_library")

package(default_visibility = ["//visibility:public"])

cc_library(
    name = "core",
    srcs = glob(["core/**/*.cc"], allow_empty = True),
    hdrs = glob(["core/**/*.h", "core/**/*.hpp"], allow_empty = True),
)
