package cmd

import (
	"fmt"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
)

var autodetectCmd = &cobra.Command{
	Use:   "autodetect [path]",
	Short: "Detect project information",
	Long:  `Scan a directory and detect programming languages, frameworks, and Docker configuration`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		fmt.Printf("Scanning directory: %s\n", targetDir)

		// Execute detection in target directory
		err := helpers.WithWorkingDir(targetDir, func() error {
			// Detect project information
			info := DetectProjectInfo()
			PrintDetectedInfo(info)
			return nil
		})

		if err != nil {
			helpers.ExitOnError(err, "failed to scan directory")
		}
	},
}
