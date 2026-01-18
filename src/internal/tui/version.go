package tui

// currentVersion is the current version of HSM (Hytale Server Manager)
// This is updated automatically by the release script
const currentVersion = "v0.1.1"

// GetVersion returns the current version string
func GetVersion() string {
	return currentVersion
}
