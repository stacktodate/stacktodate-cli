package globalconfig

import (
	"fmt"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication configuration",
	Long:  `Display information about where your authentication token is stored and its status.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, isSecure, err := helpers.GetTokenSource()

		if err != nil {
			fmt.Println("Status: Not configured")
			fmt.Println("")
			fmt.Println("To set up authentication, run:")
			fmt.Println("  stacktodate global-config set")
			return
		}

		fmt.Println("Status: Configured")
		fmt.Printf("Source: %s\n", source)

		if !isSecure {
			fmt.Println("")
			fmt.Println("⚠️  Warning: Token stored in plain text file")
			fmt.Println("For better security, use a system with OS keychain support")
		}
	},
}
