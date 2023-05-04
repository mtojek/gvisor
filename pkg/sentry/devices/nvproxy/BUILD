load("//tools:defs.bzl", "go_library")

package(default_applicable_licenses = ["//:license"])

licenses(["notice"])

go_library(
    name = "nvproxy",
    srcs = [
        "frontend.go",
        "frontend_unsafe.go",
        "nvproxy.go",
        "nvproxy_unsafe.go",
        "seccomp_filters.go",
        "uvm.go",
        "uvm_unsafe.go",
    ],
    visibility = ["//pkg/sentry:internal"],
    deps = [
        "//pkg/abi/linux",
        "//pkg/abi/nvgpu",
        "//pkg/context",
        "//pkg/errors/linuxerr",
        "//pkg/hostarch",
        "//pkg/marshal",
        "//pkg/seccomp",
        "//pkg/sentry/arch",
        "//pkg/sentry/fsimpl/devtmpfs",
        "//pkg/sentry/kernel",
        "//pkg/sentry/vfs",
        "//pkg/usermem",
        "@org_golang_x_sys//unix:go_default_library",
    ],
)
