package config

import "runtime/debug"

// Version is the git commit the server was built from. The deploy sets it with
// -ldflags "-X github.com/ArcCS/Nevermore/config.Version=<hash>" (the module
// path from go.mod); building a single file (go build ./server/server.go)
// skips Go's own VCS stamping.
var Version = ""

// GetVersion returns the build's commit hash, falling back to Go's embedded
// VCS info (package builds like go build ./server), then "unknown".
func GetVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		rev, dirty := "", false
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
		if len(rev) > 7 {
			rev = rev[:7]
		}
		if rev != "" {
			if dirty {
				rev += "-dirty"
			}
			return rev
		}
	}
	return "unknown"
}
