package version

import "runtime/debug"

var VersionOverride = ""

func Version() string {
	if VersionOverride != "" {
		return VersionOverride
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" {
			return info.Main.Version
		}
	}
	return "(unknown)"
}

func Sum() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Sum != "" {
			return info.Main.Sum
		}
	}
	return "(unknown)"
}

func Path() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path != "" {
			return info.Main.Path
		}
	}
	return "(unknown)"
}

func CommitHash() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "(unknown)"
}

func CommitTime() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return "(unknown)"
}

func Dirty() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.modified" {
				return setting.Value
			}
		}
	}
	return "(unknown)"
}
