package cmd

import (
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stacktodate",
	Short: "Official CLI for Stack To Date",
	Long:  `stacktodate - Track technology lifecycle statuses and plan for end-of-life upgrades`,
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(autodetectCmd)
}
