// go.mod for the native (non-Bazel) test harness.
//
// The test files under this directory exercise the host packages in src/go.
// Under Bazel each go_test target pins its own deps, so this file is only
// needed for running `go test` directly. The replace lets `originlang/host/...`
// imports resolve to the sources in src/go.
module originlang/tests

go 1.22

require originlang v0.0.0

replace originlang => ../../src/go