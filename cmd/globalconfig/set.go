package globalconfig

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set up authentication token",
	Long:  `Set up your stacktodate API token for authentication.\n\nThe token will be securely stored in your system's keychain or credential store.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := promptForToken()
		if err != nil {
			helpers.ExitOnError(err, "failed to read token")
		}

		if token == "" {
			helpers.ExitOnError(fmt.Errorf("token cannot be empty"), "")
		}

		// Store the token
		if err := helpers.SetToken(token); err != nil {
			helpers.ExitOnError(err, "")
		}

		source, _, _ := helpers.GetTokenSource()
		fmt.Printf("✓ Token successfully configured\n")
		fmt.Printf("  Storage: %s\n", source)
	},
}

// promptForToken prompts the user for their API token without echoing it to the terminal
func promptForToken() (string, error) {
	fmt.Print("Enter your stacktodate API token: ")

	// Read password without echoing
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	fmt.Println() // Print newline after hidden input

	token := strings.TrimSpace(string(bytePassword))
	return token, nil
}
