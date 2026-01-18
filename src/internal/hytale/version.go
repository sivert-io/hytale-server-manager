package hytale

import "fmt"

// Version information
// This should be set at build time using -ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns a full version string with commit and build date
func GetFullVersion() string {
	if Version == "dev" {
		return "dev (development build)"
	}
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}

// GetShortVersion returns a short version string for display
func GetShortVersion() string {
	if Version == "dev" {
		return "dev"
	}
	return Version
}
