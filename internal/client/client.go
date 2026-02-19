package client

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/clerk/clerk-sdk-go/v2"
)

// ClerkClient wraps Clerk API access for both the Platform API (workspace-level)
// and Backend API (per-instance, per-environment).
//
// The Platform API key is used for workspace-level operations like managing
// applications. Backend API keys are registered per application/environment
// and used for instance-level operations like managing organizations and
// environment settings.
type ClerkClient struct {
	// PlatformAPIKey is the workspace-level API key for the Clerk Platform API.
	PlatformAPIKey string

	// PlatformHTTPClient is the HTTP client used for Platform API calls.
	PlatformHTTPClient *http.Client

	// mu protects the backendClients map.
	mu sync.RWMutex

	// backendClients maps "{app_id}/{environment}" to a configured Backend API client config.
	backendClients map[string]*clerk.ClientConfig
}

// NewClerkClient creates a new ClerkClient with the given Platform API key.
func NewClerkClient(platformAPIKey string) *ClerkClient {
	return &ClerkClient{
		PlatformAPIKey:     platformAPIKey,
		PlatformHTTPClient: &http.Client{},
		backendClients:     make(map[string]*clerk.ClientConfig),
	}
}

// RegisterBackendClient registers a Backend API client for a specific
// application and environment combination. The secret key is the Backend
// API secret key for that instance.
func (c *ClerkClient) RegisterBackendClient(appID, environment, secretKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := backendClientKey(appID, environment)
	config := &clerk.ClientConfig{}
	config.Key = clerk.String(secretKey)
	c.backendClients[key] = config
}

// GetBackendConfig returns the Backend API client configuration for the given
// application and environment. Returns an error if no client is registered.
func (c *ClerkClient) GetBackendConfig(appID, environment string) (*clerk.ClientConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := backendClientKey(appID, environment)
	config, ok := c.backendClients[key]
	if !ok {
		return nil, fmt.Errorf("no backend client registered for application %q environment %q", appID, environment)
	}
	return config, nil
}

// backendClientKey returns the map key for a given app/environment pair.
func backendClientKey(appID, environment string) string {
	return appID + "/" + environment
}
