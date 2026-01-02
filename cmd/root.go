package cmd

import (
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/cmd/globalconfig"
	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/versioncheck"
	"github.com/stacktodate/stacktodate-cli/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stacktodate",
	Short: "Official CLI for Stack To Date",
	Long:  `stacktodate - Track technology lifecycle statuses and plan for end-of-life upgrades`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Only check on specific commands that should trigger automatic checks
		cmdName := cmd.Name()
		if shouldAutoCheck(cmdName) {
			showCachedUpdateNotification()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ver, _ := cmd.Flags().GetBool("version")
		if ver {
			fmt.Println(version.GetFullVersion())
			return
		}
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		helpers.ExitWithError(1, "%v", err)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(autodetectCmd)
	rootCmd.AddCommand(globalconfig.GlobalConfigCmd)
}

// shouldAutoCheck determines if a command should trigger automatic version checks
func shouldAutoCheck(cmdName string) bool {
	// Commands that should trigger automatic version checks
	autoCheckCommands := map[string]bool{
		"init":       true,
		"update":     true,
		"check":      true,
		"push":       true,
		"autodetect": true,
	}

	return autoCheckCommands[cmdName]
}

// showCachedUpdateNotification shows an update notification if cache indicates a new version is available
// This only checks the cache (no network calls) to avoid any performance impact
func showCachedUpdateNotification() {
	// Skip if update checking is disabled
	if os.Getenv("STD_DISABLE_VERSION_CHECK") == "1" {
		return
	}

	// Only check if cache is valid (don't fetch from network)
	if !versioncheck.IsCacheValid() {
		return
	}

	// Load cache
	cache, err := versioncheck.LoadCache()
	if err != nil || cache == nil {
		return
	}

	// Compare versions
	current := version.GetVersion()
	isNewer, err := versioncheck.CompareVersions(current, cache.LatestVersion)
	if err != nil || !isNewer {
		return
	}

	// Show simple notification (don't disrupt command output)
	fmt.Fprintf(os.Stderr, "\nA new version of stacktodate is available: %s → %s\n", current, cache.LatestVersion)
	fmt.Fprintf(os.Stderr, "Run 'stacktodate version --check-updates' for upgrade instructions.\n\n")
}
