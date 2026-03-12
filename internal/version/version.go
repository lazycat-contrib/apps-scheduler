// Package version provides version information for the application.
// Version is set via ldflags during build time.
//
// Example build command:
//
//	go build -ldflags "-X apps-scheduler/internal/version.Version=1.0.1" ./cmd/apps-scheduler
package version

import (
	"fmt"
	"runtime"
)

// Version is the application version.
// This value is set via ldflags during build time.
// If not set, defaults to "dev".
var Version = "dev"

// GitCommit is the git commit hash.
// This value can be set via ldflags during build time.
var GitCommit = ""

// BuildTime is the build timestamp.
// This value can be set via ldflags during build time.
var BuildTime = ""

// Info contains version information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit,omitempty"`
	BuildTime string `json:"buildTime,omitempty"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the current version info.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string.
func String() string {
	if GitCommit != "" {
		return fmt.Sprintf("%s (commit: %s)", Version, GitCommit)
	}
	return Version
}