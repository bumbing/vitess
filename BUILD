load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

exports_files([
    "data/test",
    "vendor/vendor.json",
])

# Run this rule to generate bazel build files based on the go files and imports.
# Usage:
# $ bazel run :gazelle
#
# You'll need to make the go_repos.bzl file first.
# The easiest thing to do is just run ./bazel_bootstrap.sh instead of running
# this directly.
gazelle(
    name = "gazelle",
    command = "fix",
    prefix = "vitess.io/vitess",
)

filegroup(
    name = "testdata",
    srcs = glob(["data/test/**"]),
    visibility = ["//go/testfiles:__pkg__"],
)

go_library(
    name = "go_default_library",
    srcs = ["test.go"],
    importpath = "vitess.io/vitess",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "vitess",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

pkg_tar(
    name = "binaries_dist",
    srcs = [
        "//go/cmd/mysqlctl",
        "//go/cmd/mysqlctld",
        "//go/cmd/pinschema",
        "//go/cmd/query_analyzer",
        "//go/cmd/topo2topo",
        "//go/cmd/vtclient",
        "//go/cmd/vtcombo",
        "//go/cmd/vtctl",
        "//go/cmd/vtctlclient",
        "//go/cmd/vtctld",
        "//go/cmd/vtexplain",
        "//go/cmd/vtgate",
        "//go/cmd/vtqueryserver",
        "//go/cmd/vttablet",
        "//go/cmd/vtworker",
        "//go/cmd/vtworkerclient",
        "//go/cmd/zk",
        "//go/cmd/zkctl",
        "//go/cmd/zkctld",
    ],
    mode = "0755",
    package_dir = "bin",
)

pkg_tar(
    name = "config_dist",
    srcs = glob([
        "config/**",
        "teletraan/**",
        "web/**",
    ]),
    strip_prefix = "/",
)

pkg_tar(
    name = "scripts_dist",
    srcs = ["scripts/vtgate_startup.sh"],
    mode = "0755",
    strip_prefix = "/",
)

pkg_tar(
    name = "full_dist",
    extension = "tar.gz",
    deps = [
        ":binaries_dist",
        ":config_dist",
        ":scripts_dist",
    ],
)
