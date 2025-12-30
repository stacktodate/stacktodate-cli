package version

// These variables are set via ldflags during build
var (
	Version   = "dev"      // Set by GoReleaser
	Commit    = "none"     // Git commit hash
	Date      = "unknown"  // Build date
	BuiltBy   = "unknown"  // Builder (goreleaser)
)

// GetVersion returns the version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns detailed version information
func GetFullVersion() string {
	return "stacktodate " + Version + " (commit: " + Commit + ", built: " + Date + ")"
}
