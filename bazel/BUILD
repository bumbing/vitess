# A script that generates go_repos.bzl, which will contain a bazel-compatible
# version of the govendor dependency list from vendor/vendor.json
py_binary(
    name = "govendor_to_bazel",
    srcs = ["govendor_to_bazel.py"],
)

# Generates the go_repos.bzl file using govendor_to_bazel.
genrule(
    name = "generate_go_repos",
    srcs = ["//:vendor/vendor.json"],
    outs = ["go_repos.bzl"],
    cmd = "$(location :govendor_to_bazel) $(location //:vendor/vendor.json) > $@",
    tools = [":govendor_to_bazel"],
)
