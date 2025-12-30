package cmd

import (
	"fmt"
	"os"

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

		// Change to target directory for detection
		originalDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		if targetDir != "." {
			if err := os.Chdir(targetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error changing to directory %s: %v\n", targetDir, err)
				os.Exit(1)
			}
			defer os.Chdir(originalDir)
		}

		fmt.Printf("Scanning directory: %s\n", targetDir)

		// Detect project information
		info := DetectProjectInfo()
		PrintDetectedInfo(info)
	},
}
