package piperutils

import "runtime/debug"

func GetVersion() string {
	if build, ok := debug.ReadBuildInfo(); ok && build != nil {
		return build.Main.Version
	}
	return "n/a"
}
