// Package common holds what the command, the wasm build and the docs share:
// the program's name and its version.
package version

import "runtime/debug"

// AppName is the command's name.
const AppName = "algorithm"

// Version is set at link time (-ldflags "-X .../internal/version.Version=v1.2.3").
// Without it, the module version from the build info is used, so
// `go install ...@vX.Y.Z` reports the tag.
var Version = ""

// Get returns Version, the build info's module version, or "dev".
func Get() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
