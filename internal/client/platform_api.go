package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const platformAPIBaseURL = "https://api.clerk.com/v1"

// PlatformApplicationInstance represents an instance (dev/prod) within a Clerk application.
type PlatformApplicationInstance struct {
	InstanceID      string `json:"instance_id"`
	EnvironmentType string `json:"environment_type"`
	PublishableKey  string `json:"publishable_key"`
	SecretKey       string `json:"secret_key,omitempty"`
}

// PlatformApplicationResponse is the response from the Platform API for application operations.
type PlatformApplicationResponse struct {
	ApplicationID string                        `json:"application_id"`
	Instances     []PlatformApplicationInstance `json:"instances"`
}

// PlatformCreateApplicationRequest is the request body for creating an application.
type PlatformCreateApplicationRequest struct {
	Name             string   `json:"name"`
	Domain           string   `json:"domain,omitempty"`
	ProxyPath        string   `json:"proxy_path,omitempty"`
	EnvironmentTypes []string `json:"environment_types,omitempty"`
	Template         string   `json:"template,omitempty"`
}

// PlatformUpdateApplicationRequest is the request body for updating an application.
type PlatformUpdateApplicationRequest struct {
	Name string `json:"name,omitempty"`
}

// PlatformDeletedObjectResponse is the response from the Platform API for delete operations.
type PlatformDeletedObjectResponse struct {
	Deleted bool   `json:"deleted"`
	Object  string `json:"object"`
	ID      string `json:"id"`
}

// PlatformAPIError represents an error response from the Clerk Platform API.
type PlatformAPIError struct {
	StatusCode int
	Body       string
}

func (e *PlatformAPIError) Error() string {
	return fmt.Sprintf("clerk platform API error (status %d): %s", e.StatusCode, e.Body)
}

// CreateApplication creates a new Clerk application via the Platform API.
func (c *ClerkClient) CreateApplication(ctx context.Context, req PlatformCreateApplicationRequest) (*PlatformApplicationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling create request: %w", err)
	}

	resp, err := c.platformRequest(ctx, http.MethodPost, "/platform/applications", body, nil)
	if err != nil {
		return nil, err
	}

	var result PlatformApplicationResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling create response: %w", err)
	}
	return &result, nil
}

// GetApplication retrieves a Clerk application by ID via the Platform API.
// If includeSecretKeys is true, the response will include secret keys for each instance.
func (c *ClerkClient) GetApplication(ctx context.Context, applicationID string, includeSecretKeys bool) (*PlatformApplicationResponse, error) {
	var query map[string]string
	if includeSecretKeys {
		query = map[string]string{"include_secret_keys": "true"}
	}

	resp, err := c.platformRequest(ctx, http.MethodGet, "/platform/applications/"+applicationID, nil, query)
	if err != nil {
		return nil, err
	}

	var result PlatformApplicationResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling get response: %w", err)
	}
	return &result, nil
}

// UpdateApplication updates a Clerk application by ID via the Platform API.
func (c *ClerkClient) UpdateApplication(ctx context.Context, applicationID string, req PlatformUpdateApplicationRequest) (*PlatformApplicationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling update request: %w", err)
	}

	resp, err := c.platformRequest(ctx, http.MethodPatch, "/platform/applications/"+applicationID, body, nil)
	if err != nil {
		return nil, err
	}

	var result PlatformApplicationResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling update response: %w", err)
	}
	return &result, nil
}

// DeleteApplication deletes a Clerk application by ID via the Platform API.
func (c *ClerkClient) DeleteApplication(ctx context.Context, applicationID string) error {
	resp, err := c.platformRequest(ctx, http.MethodDelete, "/platform/applications/"+applicationID, nil, nil)
	if err != nil {
		return err
	}

	var result PlatformDeletedObjectResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("unmarshaling delete response: %w", err)
	}
	if !result.Deleted {
		return fmt.Errorf("application %s was not deleted", applicationID)
	}
	return nil
}

// ListApplications lists all Clerk applications via the Platform API.
func (c *ClerkClient) ListApplications(ctx context.Context, includeSecretKeys bool) ([]PlatformApplicationResponse, error) {
	var query map[string]string
	if includeSecretKeys {
		query = map[string]string{"include_secret_keys": "true"}
	}

	resp, err := c.platformRequest(ctx, http.MethodGet, "/platform/applications", nil, query)
	if err != nil {
		return nil, err
	}

	var result []PlatformApplicationResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling list response: %w", err)
	}
	return result, nil
}

// platformRequest executes an authenticated HTTP request against the Clerk Platform API.
func (c *ClerkClient) platformRequest(ctx context.Context, method, path string, body []byte, query map[string]string) ([]byte, error) {
	url := platformAPIBaseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.PlatformAPIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.PlatformHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &PlatformAPIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	return respBody, nil
}
