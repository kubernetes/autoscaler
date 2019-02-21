package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test"
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "net.go",
        "ipnet.go"
    ],
    importpath = "k8s.io/kubernetes/pkg/util/net",
)

go_test(
    name = "go_default_test",
    srcs = [
        "net_test.go",
        "ipnet_test.go"
    ],
    embed = [":go_default_library"],
)
