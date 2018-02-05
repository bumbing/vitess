#!/bin/bash
./govendor_to_bazel.py > go_repos.bzl
bazel run :gazelle
bazel build //go/cmd/...
