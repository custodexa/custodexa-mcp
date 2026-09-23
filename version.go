package main

import (
	"runtime/debug"
	"strings"
)

// version is set at release time with -ldflags "-X main.version=<version>".
var version string

// resolveVersion prefers the release-time value, then the module version
// recorded by the Go toolchain, and reports it without a leading "v" so every
// install path of the same release reports the same string.
func resolveVersion(injected, module string) string {
	v := injected
	if v == "" {
		v = module
	}
	if v == "" || v == "(devel)" {
		return "devel"
	}
	return strings.TrimPrefix(v, "v")
}

func buildVersion() string {
	module := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		module = info.Main.Version
	}
	return resolveVersion(version, module)
}
