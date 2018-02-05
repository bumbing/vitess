load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_binary", "go_library")

exports_files(["data/test"])

gazelle(
    name = "gazelle",
    command = "fix",
    prefix = "github.com/youtube/vitess",
)

filegroup(
    name = "testdata",
    srcs = glob(["data/test/**"]),
    visibility = ["//go/testfiles:__pkg__"],
)

go_library(
    name = "go_default_library",
    srcs = ["test.go"],
    importpath = "github.com/youtube/vitess",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "vitess",
    embed = [":go_default_library"],
    importpath = "github.com/youtube/vitess",
    visibility = ["//visibility:public"],
)
