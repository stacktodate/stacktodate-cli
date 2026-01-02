package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stacktodate/stacktodate-cli/cmd/lib/cache"
)

// Component represents a single technology in the stack
type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ConvertStackToComponents converts the detected stack format to API component format
func ConvertStackToComponents(stack map[string]StackEntry) []Component {
	components := make([]Component, 0)

	for name, entry := range stack {
		components = append(components, Component{
			Name:    name,
			Version: entry.Version,
		})
	}

	return components
}

// TechStackRequest is used for POST /api/tech_stacks
type TechStackRequest struct {
	TechStack struct {
		Name       string      `json:"name"`
		Components []Component `json:"components"`
	} `json:"tech_stack"`
}

// TechStackResponse is the response from both GET and POST tech stack endpoints
type TechStackResponse struct {
	Success   bool   `json:"success,omitempty"`
	Message   string `json:"message,omitempty"`
	TechStack struct {
		ID         string      `json:"id"`
		Name       string      `json:"name"`
		Components []Component `json:"components"`
	} `json:"tech_stack"`
}

// CreateTechStack creates a new tech stack on the API
// Returns the newly created tech stack with UUID
func CreateTechStack(token, name string, components []Component) (*TechStackResponse, error) {
	apiURL := cache.GetAPIURL()
	url := fmt.Sprintf("%s/api/tech_stacks", apiURL)

	request := TechStackRequest{}
	request.TechStack.Name = name
	request.TechStack.Components = components

	var response TechStackResponse
	if err := makeAPIRequest("POST", url, token, request, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	if response.TechStack.ID == "" {
		return nil, fmt.Errorf("API response missing project ID")
	}

	return &response, nil
}

// GetTechStack retrieves an existing tech stack from the API by UUID
// This validates that the project exists and returns its details
func GetTechStack(token, uuid string) (*TechStackResponse, error) {
	apiURL := cache.GetAPIURL()
	url := fmt.Sprintf("%s/api/tech_stacks/%s", apiURL, uuid)

	var response TechStackResponse
	if err := makeAPIRequest("GET", url, token, nil, &response); err != nil {
		return nil, err
	}

	if response.TechStack.ID == "" {
		return nil, fmt.Errorf("API response missing project ID")
	}

	return &response, nil
}

// makeAPIRequest is a private helper that handles common API request logic
func makeAPIRequest(method, url, token string, requestBody interface{}, response interface{}) error {
	var req *http.Request
	var err error

	// Create request with body if provided
	if requestBody != nil {
		requestBodyJSON, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(requestBodyJSON))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

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
		return fmt.Errorf("failed to connect to StackToDate API: %w\n\nPlease check your internet connection and try again", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses first
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid or expired token\n\nPlease update your token with: stacktodate global-config set")
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project not found: UUID does not exist\n\nPlease check the UUID or create a new project")
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var errResp struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("validation error: %s", errResp.Message)
		}
		return fmt.Errorf("validation error: the server rejected your request")
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("StackToDate API is experiencing issues (status %d)\n\nPlease try again later", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	return nil
}
