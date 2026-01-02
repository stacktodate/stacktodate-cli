package globalconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove stored authentication token",
	Long:  `Remove your stored authentication token from keychain or credential storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Confirm deletion
		source, _, _ := helpers.GetTokenSource()
		if source == "not configured" {
			fmt.Println("No credentials to delete")
			return
		}

		fmt.Printf("This will remove your token from: %s\n", source)
		fmt.Print("Are you sure you want to delete your credentials? (type 'yes' to confirm): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			helpers.ExitOnError(err, "failed to read input")
		}

		response = strings.TrimSpace(response)
		if response != "yes" {
			fmt.Println("Cancelled - credentials not deleted")
			return
		}

		// Delete the token
		if err := helpers.DeleteToken(); err != nil {
			helpers.ExitOnError(err, "")
		}

		fmt.Println("✓ Credentials deleted successfully")
	},
}
