git_repository(
    name = "io_bazel_rules_go",
    tag = "0.9.0",
    # commit = "6b653a809d768269dedffe17a66c684dbbdf13fb",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

go_rules_dependencies()
go_register_toolchains()

load(":go_repos.bzl", "register_go_repos")

register_go_repos()
