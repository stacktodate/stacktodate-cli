package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/cache"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the tech stack in your browser",
	Long:  `Open the project's tech stack page on StackToDate in your default web browser`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config to get UUID
		config, err := helpers.LoadConfigWithDefaults(configFile, true)
		if err != nil {
			helpers.ExitOnError(err, "failed to load config")
		}

		// Get API URL
		apiURL := cache.GetAPIURL()

		// Build the tech stack URL
		url := fmt.Sprintf("%s/tech_stacks/%s", apiURL, config.UUID)

		// Open in default browser
		if err := openBrowser(url); err != nil {
			helpers.ExitOnError(err, "failed to open browser")
		}

		fmt.Printf("✓ Opening %s in your browser\n", url)
	},
}

// openBrowser opens a URL in the default browser for the current operating system
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("open", url)
	case "windows":
		// Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	case "linux":
		// Linux - try xdg-open first, then fall back to others
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to stacktodate.yml config file (default: stacktodate.yml)")
}
