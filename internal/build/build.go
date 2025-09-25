package build

import (
	"runtime"

	"github.com/loozhengyuan/grench/build"
)

var (
	Version    = "v0.0.0"
	CommitHash = "dev"
	Timestamp  = "1970-01-01T00:00:00Z"
)

func Info(name string) build.Info {
	return build.Info{
		App:       name,
		System:    runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   Version,
		Commit:    CommitHash,
		Timestamp: Timestamp,
	}
}
