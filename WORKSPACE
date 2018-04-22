git_repository(
    name = "io_bazel_rules_go",
    tag = "0.11.0",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

git_repository(
    name = "bazel_gazelle",
    tag = "0.11.0",
    remote = "https://github.com/bazelbuild/bazel-gazelle.git",
)

load("//bazel:go_repos.bzl", "register_go_repos")

# NOTE(dweitzman): It's important to put this before registering bazel rules / gazelle deps
# because we may need a more recent version of grpc or something
# like that.
register_go_repos()

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()
