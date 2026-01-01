package cmd

import (
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/cmd/lib/installer"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/versioncheck"
	"github.com/stacktodate/stacktodate-cli/internal/version"
	"github.com/spf13/cobra"
)

var checkUpdates bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of stacktodate`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GetFullVersion())

		if checkUpdates {
			checkForUpdates(true)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&checkUpdates, "check-updates", false, "Check for newer versions available")
}

// checkForUpdates checks for a newer version and displays update information
func checkForUpdates(verbose bool) {
	latest, releaseURL, err := versioncheck.GetLatestVersion()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Unable to check for updates: %v\n", err)
		}
		return
	}

	current := version.GetVersion()
	isNewer, err := versioncheck.CompareVersions(current, latest)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Unable to compare versions: %v\n", err)
		}
		return
	}

	if isNewer {
		installMethod := installer.DetectInstallMethod()
		instructions := installer.GetUpgradeInstructions(installMethod, latest)

		if verbose {
			fmt.Printf("\n%s\n", formatUpdateMessage(current, latest, releaseURL, instructions))
		} else {
			// Silent notification for automatic checks
			fmt.Fprintf(os.Stderr, "\nA new version of stacktodate is available: %s → %s\n", current, latest)
			fmt.Fprintf(os.Stderr, "Run 'stacktodate version --check-updates' for upgrade instructions.\n\n")
		}
	} else if verbose {
		fmt.Println("\nYou are using the latest version.")
	}
}

// formatUpdateMessage creates a formatted update notification message
func formatUpdateMessage(current, latest, releaseURL, instructions string) string {
	return fmt.Sprintf(`
Update Available
================

Current version: %s
Latest version:  %s

%s

Release notes: %s
`, current, latest, instructions, releaseURL)
}
