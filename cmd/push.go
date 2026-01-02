package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/cache"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PushRequest struct {
	Components []Component `json:"components"`
}

type PushResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	TechStack struct {
		ID         string        `json:"id"`
		Name       string        `json:"name"`
		Components []Component   `json:"components"`
	} `json:"tech_stack"`
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push tech stack components to the API",
	Long:  `Push the components defined in stacktodate.yml to the remote API`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config with UUID validation
		config, err := helpers.LoadConfigWithDefaults(configFile, true)
		if err != nil {
			helpers.ExitOnError(err, "failed to load config")
		}

		// Get token from credentials (env var, keychain, or file)
		token, err := helpers.GetToken()
		if err != nil {
			helpers.ExitOnError(err, "")
		}

		// Get API URL from environment or use default
		apiURL := cache.GetAPIURL()

		// Convert stack to components
		components := convertStackToComponents(config.Stack)

		// Create request
		request := PushRequest{
			Components: components,
		}

		// Make API call
		if err := pushToAPI(apiURL, config.UUID, token, request); err != nil {
			helpers.ExitOnError(err, "failed to push to API")
		}

		fmt.Printf("✓ Successfully pushed %d components\n", len(components))
	},
}

func convertStackToComponents(stack map[string]helpers.StackEntry) []Component {
	var components []Component

	for name, entry := range stack {
		components = append(components, Component{
			Name:    name,
			Version: entry.Version,
		})
	}

	return components
}

func pushToAPI(apiURL, techStackID, token string, request PushRequest) error {
	// Build URL
	url := fmt.Sprintf("%s/api/tech_stacks/%s/components", apiURL, techStackID)

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response PushResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API returned success=false: %s", response.Message)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to stacktodate.yml config file (default: stacktodate.yml)")
}
